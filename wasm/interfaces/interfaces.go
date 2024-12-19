package interfaces

import (
	"github.com/bluele/gcache"
	"github.com/hashicorp/go-hclog"
)

type Engine interface {
	Name() string
	Init(logger hclog.Logger, moduleCache gcache.Cache)
	InstantiateModule(modulePath string) (WasmInstance, error)
	PrePopulateCache(modulesDir string) (int, error)
}

type WasmInstance interface {
	CallFunc(funcName string, args ...interface{}) (interface{}, error)
	GetMemoryRange(start int32, end int32) []byte
	Stop()
}
