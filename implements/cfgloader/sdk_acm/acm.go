package sdk_acm

import (
	"context"
	"net/http"
	"time"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"

	"git.bestfulfill.tech/devops/go-core/implements/httpreq"
	"git.bestfulfill.tech/devops/go-core/interfaces/ierror"
	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/kits/kerror/errorx"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type Client interface {
	// 获取配置内容
	GetConfig(ctx context.Context, data ConfigData) (content []byte, err error)
	// 获取配置原始内容（上面的接口可能会完成关联加载）
	GetOriginConfig(ctx context.Context, data ConfigData) (content []byte, err error)
	// 更新配置内容
	UpdateConfig(ctx context.Context, data ConfigData) (err error)
	// Deprecated: 已废弃
	UpdateConfigSet(ctx context.Context, data ConfigSet) (err error)
}

type (
	client struct {
		breaker httpreq.HttpBreaker

		cli ihttp.Client
	}

	ConfigSet struct {
		GroupID  string   `json:"groupID"`  // 组ID
		DataIDs  []string `json:"dataIDs"`  // 数据ID数组
		TenantID string   `json:"tenantID"` // 命名空间ID
	}

	ConfigData struct {
		TenantID  string `url:"tenantID" json:"tenantID"`
		DataID    string `url:"dataID" json:"dataID"`
		GroupID   string `url:"groupID" json:"groupID"`
		AppID     string `url:"appID" json:"appID"`         // 部署应用ID
		Namespace string `url:"namespace" json:"namespace"` // 部署命名空间
		Content   string `json:"content" url:"-"`
	}

	Ret struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	Content struct {
		Content string `json:"content"`
	}
)

type BreakerOption struct {
	Success int
	Failed  int
	Timeout time.Duration
}

type Option struct {
	BreakerOption []func(*httpreq.BreakerOption)
	CliOption     []ihttp.CliOpt
}

func WithCliOpt(opts ...ihttp.CliOpt) func(option *Option) {
	return func(option *Option) {
		option.CliOption = append(option.CliOption, opts...)
	}
}

func WithBreakerOpt(opts ...func(*httpreq.BreakerOption)) func(option *Option) {
	return func(option *Option) {
		option.BreakerOption = opts
	}
}

func initOptions(opts []func(option *Option)) *Option {
	o := &Option{
		CliOption: []ihttp.CliOpt{
			func(cli *http.Client) {
				cli.Timeout = time.Second
			},
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func NewClient(endpoint string, opts ...func(option *Option)) Client {
	o := initOptions(opts)
	cli := client{
		cli:     httpreq.NewJsonClient(endpoint, o.CliOption...),
		breaker: httpreq.NewHttpBreaker(o.BreakerOption...),
	}
	return &cli
}

const (
	configPath       = "/api/v1/config"
	configSetPath    = "/api/v1/config/set" // Deprecated
	configOriginPath = "/api/v1/config/origin"
)

// Deprecated: 已废弃
func (c *client) UpdateConfigSet(ctx context.Context, data ConfigSet) (err error) {
	resp, err := c.cli.Post(ctx, configSetPath, data)
	if err != nil {
		return
	}
	err = c.parseResp(resp, nil)
	if err != nil {
		return
	}
	return
}

func (c *client) UpdateConfig(ctx context.Context, data ConfigData) (err error) {
	resp, err := c.cli.Post(ctx, configPath, data)
	if err != nil {
		return
	}
	err = c.parseResp(resp, nil)
	return
}

func (c *client) getConfig(ctx context.Context, path string, data ConfigData) (content []byte, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err := c.cli.Get(ctx, path, data)
		if err != nil {
			logger.FromContext(ctx).WithTag("CONFIG_REQ_ERR").Errorf("get config error: %v", err)
			return
		}
		ct := new(Content)
		if err = c.parseResp(resp, ct); err != nil {
			return
		}
		content = []byte(ct.Content)
		return
	}, httpreq.WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, get config, method[%s], uri[%s]", http.MethodGet, path)
	}))
	return
}

func (c *client) GetOriginConfig(ctx context.Context, data ConfigData) (content []byte, err error) {
	return c.getConfig(ctx, configOriginPath, data)
}

func (c *client) GetConfig(ctx context.Context, data ConfigData) (content []byte, err error) {
	return c.getConfig(ctx, configPath, data)
}

func (c *client) parseResp(resp ihttp.Resp, data interface{}) (err error) {
	var ret Ret
	if resp.OK() {
		if data != nil {
			ret.Data = data
		}
		if err = resp.UnmarshalJson(&ret); err != nil {
			err = errorx.ErrInternal("invalid resp: %v content: %s", err, resp.Body().String())
			return
		}
		return
	}
	if err = resp.UnmarshalJson(&ret); err != nil {
		err = errorx.Errorf(ierror.FromHttpStatusCode(resp.Status()),
			"invalid resp: %v,status: %d,content: %s", err, resp.Status(), resp.Body().String())
		return
	}
	err = errorx.Errorf(ierror.Code(ret.Code), ret.Message)
	return
}
