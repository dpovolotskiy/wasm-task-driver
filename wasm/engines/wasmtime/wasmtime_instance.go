package wasmtime

import (
	"github.com/bytecodealliance/wasmtime-go"
	"github.com/pkg/errors"

	"huawei.com/wasm-task-driver/wasm/engines"
)

type wasmtimeInstance struct {
	store    *wasmtime.Store
	instance *wasmtime.Instance
}

func (i *wasmtimeInstance) CallFunc(funcName string, args ...interface{}) (interface{}, error) {
	moduleFunc := i.instance.GetFunc(i.store, funcName)
	if moduleFunc == nil {
		return nil, errors.Wrapf(engines.ErrNotFound, "WASM module doesn't conform calling conventions: no %s func", funcName)
	}

	funcResult, err := moduleFunc.Call(i.store, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to call function: %s", funcName)
	}

	return funcResult, nil
}

func (i *wasmtimeInstance) GetMemoryRange(start, size int32) ([]byte, error) {
	return i.instance.GetExport(i.store, "memory").Memory().UnsafeData(i.store)[start : start+size], nil
}

func (i *wasmtimeInstance) Stop() {
	i.store.Engine.IncrementEpoch()
}

func (i *wasmtimeInstance) Cleanup() {}
