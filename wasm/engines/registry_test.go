package engines

import (
	"testing"

	"github.com/bluele/gcache"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"huawei.com/wasm-task-driver/wasm/interfaces"
)

type mockedEngine struct {
	name string
}

func (e *mockedEngine) Name() string {
	return e.name
}

func (e *mockedEngine) Init(_logger hclog.Logger, _moduleCache gcache.Cache) {}

func (e *mockedEngine) InstantiateModule(_modulePath string) (interfaces.WasmInstance, error) {
	return nil, nil
}

func (e *mockedEngine) PrePopulateCache(_modulesDir string) (int, error) {
	return 0, nil
}

func TestRegister(t *testing.T) {
	tCases := []struct {
		name               string
		enginesToAdd       []interfaces.Engine
		expectedEnginesNum int
	}{
		{
			name: "add two different engines",
			enginesToAdd: []interfaces.Engine{
				&mockedEngine{
					name: "fake_engine_1",
				},
				&mockedEngine{
					name: "fake_engine_2",
				},
			},
			expectedEnginesNum: 2,
		},
		{
			name: "add two engines with same name",
			enginesToAdd: []interfaces.Engine{
				&mockedEngine{
					name: "fake_engine_1",
				},
				&mockedEngine{
					name: "fake_engine_1",
				},
			},
			expectedEnginesNum: 1,
		},
	}

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			engines = make(map[string]interfaces.Engine)

			for _, engine := range tCase.enginesToAdd {
				Register(engine)
			}

			assert.Equal(t, tCase.expectedEnginesNum, len(engines))
		})
	}
}

func TestGet(t *testing.T) {
	tCases := []struct {
		name            string
		engineToGetName string
		expectedErrMsg  string
		enginesToAdd    []interfaces.Engine
	}{
		{
			name:            "engine not found",
			engineToGetName: "absent_engine",
			expectedErrMsg:  "unable to find engine with name: absent_engine: not found",
		},
		{
			name: "successfully get engine",
			enginesToAdd: []interfaces.Engine{
				&mockedEngine{
					name: "fake_engine_1",
				},
			},
			engineToGetName: "fake_engine_1",
		},
	}

	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			engines = make(map[string]interfaces.Engine)

			for _, engine := range tCase.enginesToAdd {
				Register(engine)
			}

			_, getErr := Get(tCase.engineToGetName)

			if tCase.expectedErrMsg == "" {
				assert.Nil(t, getErr)
			} else {
				assert.Equal(t, tCase.expectedErrMsg, getErr.Error())
			}
		})
	}
}
