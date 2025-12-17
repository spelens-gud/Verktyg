package cfgloader

import (
	"context"

	"github.com/spelens-gud/Verktyg.git/implements/cfgloader/sdk_acm"
	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var _ iconfig.ConfigLoader = &acm{}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type acm struct {
	*AcmOption
}

type AcmOption = NacosOption

// 通过代理层的sdk加载配置 功能较为全面
func NewAcmLoader(opts ...func(option *AcmOption)) iconfig.ConfigLoader {
	opt := newNacosOptionFromEnv()
	for _, o := range opts {
		o(&opt)
	}
	return &acm{&opt}
}

func (c *acm) MustLoad(configRoot interface{}) { must(c.Load(configRoot)) }

func (c *acm) Load(configRoot interface{}) (err error) {
	return c.LoadContext(context.Background(), configRoot)
}

func (c *acm) LoadContext(ctx context.Context, configRoot interface{}) (err error) {
	return wrapLoad(ctx, configRoot, func(ctx context.Context) (err error) {
		if err = c.Validate(); err != nil {
			return
		}

		var (
			client = sdk_acm.NewClient(c.Endpoint)
			param  = sdk_acm.ConfigData{
				TenantID: c.TenantID,
				DataID:   c.DataID,
				GroupID:  c.GroupID,
			}
			lg = getLogger(ctx, c).WithField("param", param)

			typ iconfig.ConfigType
		)

		lg.Infof("start load config [ %s : %s ]", c.GroupID, c.DataID)

		return loadFromBytes(logger.WithContext(ctx, lg), func() ([]byte, error) {
			return client.GetConfig(ctx, param)
		}, configRoot, typ)
	})
}
