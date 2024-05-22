package wasmtime

import (
	"fmt"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/client/lib/fifo"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// taskHandle should store all relevant runtime information
// such as process ID if this is a local task or other meta
// data if this driver deals with external APIs.
type taskHandle struct {
	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	logger      hclog.Logger
	taskConfig  *drivers.TaskConfig
	procState   drivers.TaskState
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult

	modulePath       string
	ioBufferConf     IOBufferConfig
	mainFunc         Main
	wasmModulesCache gcache.Cache
	store            *wasmtime.Store
	completionCh     chan struct{}
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.procState,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return h.procState == drivers.TaskStateRunning
}

func (h *taskHandle) run() {
	defer close(h.completionCh)

	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}
	h.stateLock.Unlock()

	module, err := h.getModule()
	if err != nil {
		h.reportError(err)

		return
	}

	instance, err := wasmtime.NewInstance(h.store, module, []wasmtime.AsExtern{})
	if err != nil {
		h.reportError(fmt.Errorf("unable to create WASM instance: %w", err))

		return
	}

	var ioBuffer []byte

	if h.ioBufferConf.Enabled {
		inputByte := []byte(h.ioBufferConf.InputValue)
		if len(inputByte) > int(h.ioBufferConf.Size) {
			h.reportError(fmt.Errorf("input must be less than %d bytes to fit IO buffer", h.ioBufferConf.Size))

			return
		}

		getIOBufferPtrFunc := instance.GetFunc(h.store, h.ioBufferConf.IOBufFuncName)
		if getIOBufferPtrFunc == nil {
			h.reportError(fmt.Errorf("WASM module doesn't conform calling conventions: no get_io_buffer_ptr func"))

			return
		}

		result, err2 := getIOBufferPtrFunc.Call(h.store, intListToIfaceList(h.ioBufferConf.Args)...)
		if err2 != nil {
			h.reportError(fmt.Errorf("unable to call get_io_buffer_ptr function: %w", err2))

			return
		}

		offset := result.(int32)
		ioBuffer = instance.GetExport(h.store, "memory").Memory().UnsafeData(h.store)[offset : offset+h.ioBufferConf.Size]
		n := copy(ioBuffer, inputByte)

		h.logger.Debug("copied data from task config to IO buffer", "bytes", n)

		h.mainFunc.Args = append(h.mainFunc.Args, int32(n))
	}

	moduleMainFunc := instance.GetFunc(h.store, h.mainFunc.MainFuncName)
	if moduleMainFunc == nil {
		h.reportError(fmt.Errorf("WASM module doesn't conform calling conventions: no handle_buffer func"))

		return
	}

	result, err := moduleMainFunc.Call(h.store, intListToIfaceList(h.mainFunc.Args)...)
	if err != nil {
		h.reportError(fmt.Errorf("failed to call handle_buffer(): %w", err))

		return
	}

	var out []byte

	if h.ioBufferConf.Enabled {
		resultSize := result.(int32)
		if resultSize == 0 {
			h.reportError(fmt.Errorf("unsuccessful WASM call"))

			return
		}

		out = make([]byte, resultSize)
		_ = copy(out, ioBuffer[:resultSize])
	} else {
		out = []byte(fmt.Sprintf("%v", result))
	}

	stdio, err := fifo.OpenWriter(h.taskConfig.StdoutPath)
	if err != nil {
		h.reportError(err)

		return
	}
	defer stdio.Close()

	n, err := stdio.Write(out)
	if err != nil {
		h.reportError(err)

		return
	}

	h.logger.Debug("wrote data to stdout", "bytes", n)

	h.reportCompletion()
}

func (h *taskHandle) reportError(err error) {
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.exitResult.Err = err
	h.procState = drivers.TaskStateUnknown
	h.completedAt = time.Now()
}

func (h *taskHandle) reportCompletion() {
	h.stateLock.Lock()
	defer h.stateLock.Unlock()

	h.procState = drivers.TaskStateExited
	h.exitResult.ExitCode = 0
	h.exitResult.Signal = 0
	h.completedAt = time.Now()
}

func (h *taskHandle) stop() {
	h.store.Engine.IncrementEpoch()
}

func (h *taskHandle) getModule() (*wasmtime.Module, error) {
	var (
		err    error
		module *wasmtime.Module
	)

	if h.wasmModulesCache != nil {
		mod, getCacheErr := h.wasmModulesCache.Get(h.modulePath)
		switch getCacheErr {
		case nil:
			module, err = wasmtime.NewModuleDeserialize(h.store.Engine, mod.([]byte))
			if err != nil {
				h.logger.Error("unable to deserialize WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to deserialize WASM module: %w", err)
			}
		case gcache.KeyNotFoundError:
			module, err = wasmtime.NewModuleFromFile(h.store.Engine, h.modulePath)
			if err != nil {
				h.logger.Error("unable to load WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to load WASM module: %w", err)
			}

			serModule, err2 := module.Serialize()
			if err2 != nil {
				h.logger.Error("unable to serialize WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to serialize WASM module: %w", err)
			}

			if err2 := h.wasmModulesCache.Set(h.modulePath, serModule); err2 != nil {
				h.logger.Error("unable to cache WASM module", "error", hclog.Fmt("%+v", err2))

				return nil, fmt.Errorf("unable to cache WASM module: %w", err2)
			}

			h.logger.Debug("cached WASM module", "module", h.modulePath)
		default:
			h.logger.Error("unable to get module from cache", "error", hclog.Fmt("%+v", getCacheErr))

			return nil, fmt.Errorf("unable to cache WASM module: %w", getCacheErr)
		}
	} else {
		h.logger.Debug("modules cache disabled loading WASM module from file", "module", h.modulePath)

		module, err = wasmtime.NewModuleFromFile(h.store.Engine, h.modulePath)
		if err != nil {
			h.logger.Error("unable to load WASM module", "error", hclog.Fmt("%+v", err))

			return nil, fmt.Errorf("unable to load WASM module: %w", err)
		}
	}

	return module, nil
}

func intListToIfaceList(input []int32) []interface{} {
	result := make([]interface{}, len(input))

	for i := range input {
		result[i] = input[i]
	}

	return result
}
