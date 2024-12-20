package wasmedge

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluele/gcache"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"huawei.com/wasm-task-driver/wasm/engines"
	"huawei.com/wasm-task-driver/wasm/interfaces"
)

const engineExtensionName = "wasmedge"

func init() {
	engines.Register(&wasmedgeEngine{})
}

type wasmedgeEngine struct {
	logger       hclog.Logger
	modulesCache gcache.Cache
}

func (e *wasmedgeEngine) Name() string {
	return engineExtensionName
}

func (e *wasmedgeEngine) Init(logger hclog.Logger, moduleCache gcache.Cache) {
	e.logger = logger
	e.modulesCache = moduleCache
}

func (e *wasmedgeEngine) PrePopulateCache(modulesDir string) (int, error) {
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

	store := wasmedge.NewStore()
	defer store.Release()

	vm := wasmedge.NewVMWithStore(store)
	defer vm.Release()

	for _, modulePath := range modulesPath {
		wasmModule, err := loadModule(vm, modulePath)
		if err != nil {
			return 0, fmt.Errorf("unable to load WASM module (%v) from file: %v", modulePath, err)
		}

		if err := e.modulesCache.Set(modulePath, wasmModule); err != nil {
			return 0, fmt.Errorf("unable to cache WASM module (%v)", modulePath)
		}

		preCachedModulesNumber++

		e.logger.Trace("WASM module pre-cached", "module", modulePath)
	}

	return preCachedModulesNumber, nil
}

func (e *wasmedgeEngine) InstantiateModule(modulePath string) (interfaces.WasmInstance, error) {
	e.logger.Debug("instantiate new module", "module path", modulePath)

	store := wasmedge.NewStore()
	vm := wasmedge.NewVMWithStore(store)

	module, err := e.getModule(vm, modulePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get module %s: %w", modulePath, err)
	}

	return &wasmedgeInstance{
		module: module,
		vm:     vm,
	}, nil
}

func (e *wasmedgeEngine) getModule(vm *wasmedge.VM, modulePath string) (*wasmedge.Module, error) {
	var (
		err       error
		astModule *wasmedge.AST
	)

	if e.modulesCache != nil {
		mod, getCacheErr := e.modulesCache.Get(modulePath)

		switch {
		case getCacheErr == nil:
			astModule = mod.(*wasmedge.AST)
		case errors.Is(getCacheErr, gcache.KeyNotFoundError):
			astModule, err = loadModule(vm, modulePath)
			if err != nil {
				e.logger.Error("unable to load WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to load WASM module: %w", err)
			}

			if err = e.modulesCache.Set(modulePath, astModule); err != nil {
				e.logger.Error("unable to cache WASM module", "error", hclog.Fmt("%+v", err))

				return nil, fmt.Errorf("unable to cache: %w", err)
			}

			e.logger.Debug("cached WASM module", "module", modulePath)
		default:
			e.logger.Error("unable to get module from cache", "error", hclog.Fmt("%+v", getCacheErr))

			return nil, fmt.Errorf("unable to cache WASM module: %w", getCacheErr)
		}
	}

	wasmModule, err := vm.GetExecutor().Instantiate(vm.GetStore(), astModule)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate executor: %w", err)
	}

	return wasmModule, nil
}

func loadModule(vm *wasmedge.VM, filePath string) (*wasmedge.AST, error) {
	moduleByte, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	loader := vm.GetLoader()

	module, err := loader.LoadBuffer(moduleByte)
	if err != nil {
		return nil, fmt.Errorf("unable to load modulebyte buffer: %w", err)
	}

	validator := vm.GetValidator()

	if err := validator.Validate(module); err != nil {
		return nil, fmt.Errorf("unable to validate module: %w", err)
	}

	return module, nil
}
