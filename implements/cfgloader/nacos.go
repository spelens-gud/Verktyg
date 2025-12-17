package cfgloader

import (
	"context"
	"errors"
	"os"

	"github.com/spelens-gud/Verktyg.git/implements/cfgloader/sdk_nacos"
	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/kits/kgo"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var _ iconfig.ConfigLoader = &nacos{}

type nacos struct {
	*NacosOption
}

type NacosOption struct {
	DataID   string
	GroupID  string
	TenantID string
	Endpoint string
	Username string
	Password string
}

func (opt NacosOption) Validate() (err error) {
	if len(opt.DataID) == 0 || len(opt.GroupID) == 0 || len(opt.TenantID) == 0 || len(opt.Endpoint) == 0 {
		return errors.New("invalid loader option")
	}
	return
}

func newNacosOptionFromEnv() NacosOption {
	return NacosOption{
		DataID:   kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmDataID)),
		GroupID:  kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmGroupID)),
		TenantID: kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmTenantID)),
		Endpoint: kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmEndpoint)),
		Username: kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmAccessID)),
		Password: kgo.InitStrings(os.Getenv(iconfig.EnvKeyAcmAccessKey)),
	}
}

// 真正的nacos
func NewNacosLoader(opts ...func(option *NacosOption)) iconfig.ConfigLoader {
	opt := newNacosOptionFromEnv()
	for _, o := range opts {
		o(&opt)
	}
	return &nacos{&opt}
}

func (c *nacos) MustLoad(configRoot interface{}) { must(c.Load(configRoot)) }

func (c *nacos) Load(configRoot interface{}) (err error) {
	return c.LoadContext(context.Background(), configRoot)
}

func (c *nacos) LoadContext(ctx context.Context, configRoot interface{}) (err error) {
	return wrapLoad(ctx, configRoot, func(ctx context.Context) (err error) {
		if err = c.Validate(); err != nil {
			return
		}

		var (
			client = sdk_nacos.NewClient(sdk_nacos.NacosConfig{
				AuthConfig: sdk_nacos.AuthConfig{
					Username: c.Username,
					Password: c.Password,
				},
				Endpoint: c.Endpoint,
			})
			param = sdk_nacos.ConfigData{
				TenantID: c.TenantID,
				DataID:   c.DataID,
				GroupID:  c.GroupID,
			}
			lg = getLogger(ctx, c).WithField("param", param)
		)

		lg.Infof("start load config [ %s : %s ]", c.GroupID, c.DataID)

		return loadFromBytes(logger.WithContext(ctx, lg), func() ([]byte, error) {
			return client.GetRefConfig(ctx, param)
		}, configRoot, param.Type)
	})
}
