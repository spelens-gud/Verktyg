package cfgloader

import (
	"context"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
)

var _ iconfig.ConfigLoader = multi{}

type multi []iconfig.ConfigLoader

func (m multi) MustLoad(configRoot interface{}) { must(m.Load(configRoot)) }

func (m multi) Load(configRoot interface{}) (err error) {
	return m.LoadContext(context.Background(), configRoot)
}

func (m multi) LoadContext(ctx context.Context, configRoot interface{}) (err error) {
	return wrapLoad(ctx, configRoot, func(ctx context.Context) (err error) {
		for _, loader := range m {
			if err = loader.LoadContext(ctx, configRoot); err == nil {
				return
			}
		}
		getLogger(ctx, m).Errorf("load config from all loader failed")
		return
	})
}

// 多个loader结合
func NewMultiLoader(loaders ...iconfig.ConfigLoader) iconfig.ConfigLoader {
	return multi(loaders)
}
