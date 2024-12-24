package wasmedge

import (
	"github.com/pkg/errors"
	"github.com/second-state/WasmEdge-go/wasmedge"

	"huawei.com/wasm-task-driver/wasm/engines"
)

type wasmedgeInstance struct {
	module *wasmedge.Module
	vm     *wasmedge.VM
}

func (i *wasmedgeInstance) CallFunc(funcName string, args ...interface{}) (interface{}, error) {
	moduleFunc := i.module.FindFunction(funcName)
	if moduleFunc == nil {
		return nil, errors.Wrapf(engines.ErrNotFound, "WASM module doesn't conform calling conventions: no %s func", funcName)
	}

	funcResult, err := i.vm.GetExecutor().Invoke(moduleFunc, args...)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to call function: %s", funcName)
	}

	return funcResult[0], nil
}

func (i *wasmedgeInstance) GetMemoryRange(start, size int32) ([]byte, error) {
	memory := i.module.FindMemory("memory")

	//nolint:gosec
	ioBuf, err := memory.GetData(uint(start), uint(size))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get data of memory")
	}

	return ioBuf, nil
}

// TODO: find way to interrupt wasmedge instance execution.
func (i *wasmedgeInstance) Stop() {}

func (i *wasmedgeInstance) Cleanup() {
	defer i.vm.GetStore().Release()

	i.module.Release()
	i.vm.Release()
}
