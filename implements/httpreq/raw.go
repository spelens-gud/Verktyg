package httpreq

import (
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/kits/kgo/buffpool"
)

var _ ihttp.RawClient = &RawClient{}

type RawClient struct {
	// 已从Cli更为client 请使用Client()获取
	client *http.Client
}

func (r RawClient) Client() *http.Client {
	return r.client
}

func (r RawClient) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

func (r RawClient) Do(ctx context.Context, req *http.Request) (resp ihttp.Resp, err error) {
	if ctx != req.Context() {
		req = req.WithContext(ctx)
	}
	return do(r.client, req)
}

func (r RawClient) Get(ctx context.Context, url string) (resp ihttp.Resp, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return r.Do(ctx, req)
}

func (r RawClient) PostForm(ctx context.Context, url string, data url.Values) (resp ihttp.Resp, err error) {
	bf := buffpool.GetBytesBuffer()
	defer buffpool.PutBytesBuffer(bf)
	if _, err = bf.WriteString(data.Encode()); err != nil {
		return
	}
	return r.Post(ctx, url, ihttp.ContentTypeForm, bf)
}

func (r RawClient) Post(ctx context.Context, url, contentType string, body io.Reader) (resp ihttp.Resp, err error) {
	req, err := r.NewRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set(ihttp.HeaderContentType, contentType)
	return r.Do(ctx, req)
}

func (r RawClient) NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, body)
}

func NewRawClient(opts ...ihttp.CliOpt) ihttp.RawClient {
	cli := ihttp.NewDefaultHttpClient(opts...)
	return &RawClient{
		client: cli,
	}
}
