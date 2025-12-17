package iconfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/spelens-gud/Verktyg/kits/kgo"
	"github.com/spelens-gud/Verktyg/version"
)

const (
	ConfigTypeUnknown ConfigType = iota
	ConfigTypeJson
	ConfigTypeYaml
)

type (
	ConfigLoader interface {
		LoadContext(ctx context.Context, configRoot interface{}) (err error)
		Load(configRoot interface{}) (err error)
		MustLoad(configRoot interface{})
	}

	Env string

	ConfigType uint
)

var env = configString{
	Init: func() string {
		ret := kgo.InitStrings(os.Getenv(EnvKey), os.Getenv(envKeyOld), Development)
		fmt.Printf("[SY-CORE] env loaded [ %s ] version [ %s ]\n", ret, version.GetVersion())
		return ret
	},
}

func GetEnv() Env { return Env(env.Get()) }

func SetEnv(e string) {
	fmt.Printf("[SYCORE] env updated [ %s ]\n", e)
	env.Set(e)
}

func (env Env) String() string {
	switch env {
	case Development:
		return Development
	case Testing:
		return Testing
	case PreRelease:
		return PreRelease
	case Production:
		return Production
	default:
		return string(env)
	}
}

func (env Env) IsDevelopment() bool { return env == Development || len(env) == 0 }

func (env Env) IsTesting() bool { return env == Testing }

func (env Env) IsPreRelease() bool { return env == PreRelease }

func (env Env) IsProduction() bool { return env == Production }

func (t ConfigType) Unmarshal(data []byte, config interface{}) (err error) {
	if !GetEnv().IsProduction() {
		var bf bytes.Buffer
		if err := json.Indent(&bf, data, "", "\t"); err != nil {
			fmt.Printf("\n[CONFIG]:\n%s\n", data)
		} else {
			fmt.Printf("\n[CONFIG]:\n%s\n", bf.String())
		}
	}

	switch t {
	case ConfigTypeYaml:
		return yaml.Unmarshal(data, config)
	case ConfigTypeJson:
		return json.Unmarshal(data, config)
	// 默认JSON
	default:
		return json.Unmarshal(data, config)
	}
}
