package engines

import (
	"github.com/pkg/errors"
	"huawei.com/wasm-task-driver/wasm/interfaces"
)

var engines = make(map[string]interfaces.Engine)

func Register(engine interfaces.Engine) {
	engines[engine.Name()] = engine
}

func Get(name string) (interfaces.Engine, error) {
	var (
		engine interfaces.Engine
		found  bool
	)

	if engine, found = engines[name]; !found {
		return nil, errors.Wrapf(ErrNotFound, "unable to find engine with name: %s", name)
	}

	return engine, nil
}
