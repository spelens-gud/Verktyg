package ihttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"git.bestfulfill.tech/devops/go-core/kits/knet"
)

type (
	Client interface {
		// 获取原生方式调用接口Client
		Raw() RawClient

		// 请求方法
		Get(ctx context.Context, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)
		Post(ctx context.Context, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)
		Put(ctx context.Context, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)
		PATCH(ctx context.Context, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)
		Delete(ctx context.Context, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)

		// 请求
		Do(ctx context.Context, method, url string, req interface{}, opts ...ReqOpt) (resp Resp, err error)

		// 复制一个Client 带上指定的选项
		WithRequestOptions(opts ...ReqOpt) Client

		// 新建http Request
		NewRequest(ctx context.Context, method, reqUrl string, req interface{}) (request *http.Request, err error)
	}

	RawClient interface {
		// 获取原生*http.Client
		Client() *http.Client
		// 关闭连接池
		Close() error
		// 发起http.Request请求
		Do(ctx context.Context, r *http.Request) (resp Resp, err error)
		// 发起GET请求
		Get(ctx context.Context, url string) (resp Resp, err error)
		// 发起POST请求 content-type 为application/x-www-form-urlencoded
		PostForm(ctx context.Context, url string, data url.Values) (resp Resp, err error)
		// 发起自定义POST请求
		Post(ctx context.Context, url, contentType string, body io.Reader) (resp Resp, err error)
		// 新建请求
		NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)
	}

	ReqOpt func(r *http.Request)

	CliOpt func(cli *http.Client)

	Resp interface {
		Status() int
		UnmarshalJson(interface{}) (err error)
		Header() http.Header
		Resp() *http.Response
		Body() *bytes.Buffer
		Release()
		OK() bool
	}

	WrappedTransport interface {
		Transport() (*http.Transport, bool)
	}

	RedirectError string

	IRedirectError interface {
		error
		RedirectError()
	}
)

func (e RedirectError) RedirectError() {}

func (e RedirectError) Error() string { return string(e) }

func NewRedirectError(msg string) error { return RedirectError(msg) }

type wrapRoundTripper struct {
	r  http.RoundTripper
	do func(doRoundTrip func(*http.Request) (*http.Response, error), request *http.Request) (resp *http.Response, err error)
}

func WrapRoundTripper(r http.RoundTripper, do func(doRoundTrip func(*http.Request) (*http.Response, error), request *http.Request) (resp *http.Response, err error)) http.RoundTripper {
	if r == nil {
		r = http.DefaultTransport
	}
	return &wrapRoundTripper{r, do}
}

func (w *wrapRoundTripper) Transport() (*http.Transport, bool) {
	return UnwrapTransport(w.r)
}

func UnwrapTransport(r http.RoundTripper) (*http.Transport, bool) {
	if t, ok := r.(WrappedTransport); ok {
		return t.Transport()
	}
	tp, ok := r.(*http.Transport)
	return tp, ok
}

func (w *wrapRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return w.do(w.r.RoundTrip, request)
}

const DefaultDialKeepAlive = 120 * time.Second

func NewDefaultTransport() http.RoundTripper {
	return &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         knet.WrapTcpDialer(time.Second, DefaultDialKeepAlive, nil),
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        (runtime.GOMAXPROCS(-1) + 1) * 20,
		MaxIdleConnsPerHost: (runtime.GOMAXPROCS(-1) + 1) * 10,
		IdleConnTimeout:     DefaultDialKeepAlive,
		TLSHandshakeTimeout: 10 * time.Second,
		//ExpectContinueTimeout: 1 * time.Second,
	}
}

func NewDefaultHttpClient(opt ...CliOpt) *http.Client {
	cli := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: NewDefaultTransport(),
	}
	for _, o := range opt {
		o(cli)
	}
	return cli
}

const (
	DefaultTimeout = 5 * time.Second

	HeaderContentType   = "Content-Type"
	HeaderContentLength = "Content-Length"

	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJson = "application/json"
)

func SplitHostPort(addr string) (host string, port int, err error) {
	sph := strings.Split(addr, ":")
	if len(sph) == 1 {
		return sph[0], 80, nil
	}
	if len(sph) == 2 {
		port, err = strconv.Atoi(sph[1])
		if err != nil {
			return
		}
		return sph[0], port, err
	}
	return "", 0, errors.New("invalid addr")
}

func JoinUrl(host, reqUrl string) (u *url.URL, err error) {
	if len(host) == 0 {
		return url.Parse(reqUrl)
	}

	if len(reqUrl) == 0 {
		return url.Parse(host)
	}

	u, err = url.Parse(host)
	if err != nil {
		return
	}

	u2, err := url.Parse(reqUrl)
	if err != nil {
		return
	}

	if len(u.Path) == 0 || u.Path == "/" {
		u.Path = u2.Path
	} else {
		u.Path = path.Join(u.Path, u2.Path)
		if strings.HasSuffix(u2.Path, "/") {
			u.Path += "/"
		}
	}

	q1, _ := url.ParseQuery(u.RawQuery)
	q2, _ := url.ParseQuery(u2.RawQuery)
	for k := range q2 {
		q1.Set(k, q2.Get(k))
	}
	u.RawQuery = q1.Encode()
	return
}
