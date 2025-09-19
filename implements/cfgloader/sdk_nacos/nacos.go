package sdk_nacos

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"git.bestfulfill.tech/devops/go-core/implements/httpreq"
	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/interfaces/ierror"
	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/kits/kerror/errorx"
	"git.bestfulfill.tech/devops/go-core/kits/kjson"
)

type Client interface {
	GetRefConfig(ctx context.Context, data ConfigData) (content []byte, err error)

	GetConfig(ctx context.Context, data ConfigData) (content []byte, err error)

	UpdateConfig(ctx context.Context, data ConfigData) (err error)

	DelConfig(ctx context.Context, data ConfigData) (err error)

	ListenConfig(ctx context.Context, data ListenConfigData) (content UpdatedConfig, ok bool, err error)

	GetAccessToken(ctx context.Context) (accessToken string)
}

const (
	OKResp = "true"

	identSplitConfig      = string(rune(1))
	identSplitConfigInner = string(rune(2))

	configPath = "/nacos/v1/cs/configs"
	listenPath = "/nacos/v1/cs/configs/listener"
	authPath   = "/nacos/v1/auth/login"
	authQuery  = "accessToken=%s"
)

type (
	client struct {
		cli                ihttp.Client
		accessToken        *atomic.Value
		tokenTtl           int
		tokenRefreshWindow int
		NacosConfig
	}

	NacosConfig struct {
		AuthConfig
		Endpoint string `json:"endpoint"`
	}

	AuthConfig struct {
		Username string `json:"username" form:"username" url:"username"`
		Password string `json:"password" form:"password" url:"password"`
	}

	ConfigData struct {
		Type     iconfig.ConfigType `url:"-"`
		TenantID string             `url:"tenant"`
		DataID   string             `url:"dataId"`
		GroupID  string             `url:"group"`
		Content  string             `url:"content"`
	}

	UpdatedConfig struct {
		DataID   string
		GroupID  string
		TenantID string
	}

	ListenConfigData struct {
		TenantID string
		Configs  []ConfigData
	}

	authData struct {
		AccessToken string `json:"accessToken"`
		TokenTtl    int    `json:"tokenTtl"`
	}
)

type Option struct {
	CliOption []ihttp.CliOpt
}

func WithCliOpt(opts ...ihttp.CliOpt) func(option *Option) {
	return func(option *Option) {
		option.CliOption = append(option.CliOption, opts...)
	}
}

func initOptions(opts []func(option *Option)) *Option {
	o := &Option{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func NewClient(config NacosConfig, opts ...func(option *Option)) Client {
	o := initOptions(opts)
	cli := client{
		cli:         httpreq.NewFormClient(config.Endpoint, o.CliOption...),
		NacosConfig: config,
		accessToken: &atomic.Value{},
	}
	cli.autoRefresh(context.Background())
	return &cli
}

func (c *client) GetAccessToken(ctx context.Context) (accessToken string) {
	v := c.accessToken.Load()
	if v == nil {
		accessToken, _ = c.login(ctx)
		return
	}
	return v.(string)
}

func (c *client) autoRefresh(ctx context.Context) {
	// If the username is not set, the automatic refresh Token is not enabled
	if c.Username == "" {
		return
	}

	go func() {
		var timer *time.Timer
		if c.tokenTtl > 0 && c.tokenRefreshWindow > 0 {
			timer = time.NewTimer(time.Second * time.Duration(c.tokenTtl-c.tokenRefreshWindow))
		} else {
			timer = time.NewTimer(time.Second * time.Duration(10))
		}
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				_, err := c.login(ctx)
				if err != nil {
					timer.Reset(time.Second * time.Duration(10))
				} else {
					timer.Reset(time.Second * time.Duration(c.tokenTtl-c.tokenRefreshWindow))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *client) login(ctx context.Context) (accessToken string, err error) {
	resp, err := c.cli.Post(ctx, authPath, c.NacosConfig.AuthConfig)
	if err != nil {
		return
	}
	if !resp.OK() {
		err = ErrorHelper(resp, fmt.Sprintf("auth login %s:%s failed", c.Username, c.Password))
		return
	}
	var authData authData
	err = resp.UnmarshalJson(&authData)
	if err != nil {
		return
	}
	accessToken = authData.AccessToken
	c.tokenTtl = authData.TokenTtl
	c.tokenRefreshWindow = authData.TokenTtl / 5
	c.accessToken.Store(accessToken)
	return
}

func (c *client) makeListenReq(ctx context.Context, data ListenConfigData) (req *http.Request, err error) {
	bf := bytes.NewBufferString("Listening-Configs=")
	for _, c := range data.Configs {
		md5Hash := md5.New()
		_, _ = md5Hash.Write([]byte(c.Content))
		c.Content = hex.EncodeToString(md5Hash.Sum(nil))
		bf.WriteString(strings.Join(
			[]string{c.DataID, c.GroupID, c.Content, data.TenantID},
			identSplitConfigInner))
		bf.WriteString(identSplitConfig)
	}
	u, err := url.Parse(c.NacosConfig.Endpoint)
	if err != nil {
		return
	}
	u.Path = path.Join(u.Path, listenPath)
	q := u.Query()
	q.Set("accessToken", c.GetAccessToken(ctx))
	u.RawQuery = q.Encode()
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bf)
	if err != nil {
		return
	}
	req.Header.Add("Long-Pulling-Timeout", "30000")
	req.Header.Add(ihttp.HeaderContentType, ihttp.ContentTypeForm)
	return
}

func (c *client) makeListenResp(bs string) (content UpdatedConfig, ok bool, err error) {
	if len(bs) == 0 {
		return
	}
	bs = strings.Trim(strings.TrimSpace(bs), "%01")
	ret := strings.Split(bs, "%02")
	if len(ret) != 3 {
		err = fmt.Errorf("invalid ret :%s", bs)
		return
	}
	content.DataID = ret[0]
	content.GroupID = ret[1]
	content.TenantID = ret[2]
	ok = true
	return
}

func (c *client) ListenConfig(ctx context.Context, data ListenConfigData) (content UpdatedConfig, ok bool, err error) {
	req, err := c.makeListenReq(ctx, data)
	if err != nil {
		return
	}
	resp, err := c.cli.Raw().Do(ctx, req)
	if err != nil {
		return
	}
	if !resp.OK() {
		err = ErrorHelper(resp, fmt.Sprintf("listen config %s failed", data.TenantID))
		return
	}
	bs := resp.Body().String()
	return c.makeListenResp(bs)
}

func (c *client) GetConfig(ctx context.Context, data ConfigData) (content []byte, err error) {
	resp, err := c.cli.Get(ctx, configPath+"?"+fmt.Sprintf(authQuery, c.GetAccessToken(ctx)), data)
	if err != nil {
		return
	}
	if resp.OK() {
		content = resp.Body().Bytes()
		return
	}
	err = ErrorHelper(resp, fmt.Sprintf("get config %s:%s failed", data.GroupID, data.DataID))
	return
}

func (c *client) GetRefConfig(ctx context.Context, data ConfigData) (content []byte, err error) {
	content, err = c.GetConfig(ctx, data)
	if err != nil {
		return
	}
	// 非JSON 不关联加载
	if !json.Valid(content) {
		return
	}

	switch data.Type {
	// 暂不支持其他格式的关联KV加载
	default:
		content, err = kjson.LoadRefJson(content, func(string string) (ref []byte, err error) {
			newData := data
			if strings.ContainsRune(string, '.') {
				sp := strings.Split(string, ".")
				newData.GroupID = sp[0]
				newData.DataID = sp[1]
			} else {
				newData.DataID = string
			}
			ref, err = c.GetConfig(ctx, newData)
			return
		})
	}
	return
}

func (c *client) UpdateConfig(ctx context.Context, data ConfigData) (err error) {
	resp, err := c.cli.Post(ctx, configPath+"?"+fmt.Sprintf(authQuery, c.GetAccessToken(ctx)), data)
	if err != nil {
		return
	}
	err = ErrorHelper(resp, fmt.Sprintf("update config %s:%s failed", data.GroupID, data.DataID))
	return
}

func (c *client) DelConfig(ctx context.Context, data ConfigData) (err error) {
	resp, err := c.cli.Delete(ctx, configPath+"?"+fmt.Sprintf(authQuery, c.GetAccessToken(ctx)), data)
	if err != nil {
		return
	}
	err = ErrorHelper(resp, fmt.Sprintf("delete config %s:%s failed", data.GroupID, data.DataID))
	return
}

/*
官方文档描述：
400	Bad Request	客户端请求中的语法错误
403	Forbidden	没有权限
404	Not Found	无法找到资源
500	Internal Server Error	服务器内部错误
200	OK	正常
*/

var mapErrorDesc = map[int]string{
	http.StatusBadRequest:          "客户端请求中的语法错误",
	http.StatusForbidden:           "没有权限",
	http.StatusNotFound:            "无法找到资源",
	http.StatusInternalServerError: "服务器内部错误",
}

var mapCodeError = map[int]ierror.Code{
	http.StatusBadRequest:          ierror.InvalidArgument,
	http.StatusForbidden:           ierror.PermissionDenied,
	http.StatusNotFound:            ierror.NotFound,
	http.StatusInternalServerError: ierror.Internal,
}

func ErrorHelper(resp ihttp.Resp, msg string) (err error) {
	content := resp.Body().String()
	if content == OKResp {
		return
	}
	v, ok := mapErrorDesc[resp.Status()]
	if !ok {
		v = mapErrorDesc[http.StatusInternalServerError]
	}
	err = errorx.Errorf(mapCodeError[resp.Status()], msg+": "+v+": "+strings.TrimSpace(resp.Body().String()))
	return
}
