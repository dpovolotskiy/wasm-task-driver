package wasmtime

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/bluele/gcache"
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"huawei.com/wasm-task-driver/wasm/engines"
	"huawei.com/wasm-task-driver/wasm/interfaces"
)

const engineExtensionName = "wasmtime"

func init() {
	engines.Register(&wasmtimeEngine{})
}

type wasmtimeEngine struct {
	logger       hclog.Logger
	modulesCache gcache.Cache
}

func (e *wasmtimeEngine) Name() string {
	return engineExtensionName
}

func (e *wasmtimeEngine) Init(logger hclog.Logger, moduleCache gcache.Cache) {
	e.logger = logger
	e.modulesCache = moduleCache
}

// PrePopulateCache precache all wasm modules in specified directory
// and return number of precached modules and error.
func (e *wasmtimeEngine) PrePopulateCache(modulesDir string) (int, error) {
	if e.modulesCache == nil {
		return 0, fmt.Errorf("unable to pre populate modules: cache is not created")
	}

	var (
		modulesPath            []string
		preCachedModulesNumber int
	)

	err := filepath.Walk(modulesDir, func(path string, info fs.FileInfo, _err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".wasm") {
			modulesPath = append(modulesPath, path)
		}

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("unable to get WASM modules for pre-cache from %s directory: %v",
			modulesDir, err)
	}

	loadEngineConfig := wasmtime.NewConfig()
	loadEngineConfig.SetEpochInterruption(true)
	loadEngine := wasmtime.NewEngineWithConfig(loadEngineConfig)

	for _, modulePath := range modulesPath {
		wasmModule, err := wasmtime.NewModuleFromFile(loadEngine, modulePath)
		if err != nil {
			return 0, fmt.Errorf("unable to load WASM module (%v) from file: %v", modulePath, err)
		}

		serModule, err := wasmModule.Serialize()
		if err != nil {
			return 0, fmt.Errorf("unable to serialize WASM module (%v): %v", modulePath, err)
		}

		if err := e.modulesCache.Set(modulePath, serModule); err != nil {
			return 0, fmt.Errorf("unable to cache WASM module (%v)", modulePath)
		}

		preCachedModulesNumber++

		e.logger.Trace("WASM module pre-cached", "module", modulePath)
	}

	return preCachedModulesNumber, nil
}

func (e *wasmtimeEngine) InstantiateModule(modulePath string) (interfaces.WasmInstance, error) {
	e.logger.Debug("instantiate new module", "module path", modulePath)

	engineConfig := wasmtime.NewConfig()
	engineConfig.SetEpochInterruption(true)

	engine := wasmtime.NewEngineWithConfig(engineConfig)

	store := wasmtime.NewStore(engine)
	store.SetEpochDeadline(1)

	module, err := e.getModule(store, modulePath)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get module: %s", modulePath)
	}

	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create new instance from module: %s", modulePath)
	}

	return &wasmtimeInstance{
		store:    store,
		instance: instance,
	}, nil
}

func (e *wasmtimeEngine) getModule(store *wasmtime.Store, modulePath string) (*wasmtime.Module, error) {
	var (
		err    error
		module *wasmtime.Module
	)

	if e.modulesCache != nil {
		mod, getCacheErr := e.modulesCache.Get(modulePath)
		switch getCacheErr {
		case nil:
			module, err = wasmtime.NewModuleDeserialize(store.Engine, mod.([]byte))
			if err != nil {
				e.logger.Error("unable to deserialize WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to deserialize WASM module: %w", err)
			}
		case gcache.KeyNotFoundError:
			module, err = wasmtime.NewModuleFromFile(store.Engine, modulePath)
			if err != nil {
				e.logger.Error("unable to load WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to load WASM module: %w", err)
			}

			serModule, err2 := module.Serialize()
			if err2 != nil {
				e.logger.Error("unable to serialize WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to serialize WASM module: %w", err)
			}

			if err2 := e.modulesCache.Set(modulePath, serModule); err2 != nil {
				e.logger.Error("unable to cache WASM module", "error", hclog.Fmt("%+v", err2))

				return nil, fmt.Errorf("unable to cache WASM module: %w", err2)
			}

			e.logger.Debug("cached WASM module", "module", modulePath)
		default:
			e.logger.Error("unable to get module from cache", "error", hclog.Fmt("%+v", getCacheErr))

			return nil, fmt.Errorf("unable to cache WASM module: %w", getCacheErr)
		}
	} else {
		e.logger.Debug("modules cache disabled loading WASM module from file", "module", modulePath)

		module, err = wasmtime.NewModuleFromFile(store.Engine, modulePath)
		if err != nil {
			e.logger.Error("unable to load WASM module", "error", hclog.Fmt("%+v", err))

			return nil, fmt.Errorf("unable to load WASM module: %w", err)
		}
	}

	return module, nil
}
