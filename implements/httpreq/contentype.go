package httpreq

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spelens-gud/Verktyg/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg/kits/kgo/buffpool"
	"github.com/spelens-gud/Verktyg/kits/kurl"
)

var _ ihttp.Client = &ContentTypeClient{}

type ContentTypeClient struct {
	rawClient         ihttp.RawClient
	requestBuilderMap map[string]requestBuilder
	defaultBuilder    requestBuilder
	host              string
	requestOptions    []ihttp.ReqOpt
}

func (c *ContentTypeClient) WithRequestOptions(opts ...ihttp.ReqOpt) ihttp.Client {
	newClient := ContentTypeClient{
		rawClient:         c.rawClient,
		requestBuilderMap: c.requestBuilderMap,
		defaultBuilder:    c.defaultBuilder,
		host:              c.host,
		requestOptions:    append(c.requestOptions, opts...),
	}
	return &newClient
}

func (c *ContentTypeClient) Raw() ihttp.RawClient {
	return c.rawClient
}

type requestBuilder func(ctx context.Context, method string, u *url.URL, req interface{}) (request *http.Request, bodyBuffer *bytes.Buffer, err error)

func queryRequestBuilder(ctx context.Context, method string, u *url.URL, req interface{}) (request *http.Request, bodyBuffer *bytes.Buffer, err error) {
	// 按url/query/form tag拼接query参数 不存在tag则以字段名拼接
	// github.com/google/go-querystring/query
	query := kurl.Parse2UrlValues(req, false, "url", "query", "form")
	request, err = http.NewRequestWithContext(ctx, method, joinUrlQuery(u, query), nil)
	return
}

func NewContentTypeClient(defaultBuilder requestBuilder, builderMap map[string]requestBuilder, host string, opts ...ihttp.CliOpt) ihttp.Client {
	rawCli := NewRawClient(opts...)
	if defaultBuilder == nil {
		defaultBuilder = queryRequestBuilder
	}

	return &ContentTypeClient{
		rawClient:         rawCli,
		host:              host,
		requestBuilderMap: builderMap,
		defaultBuilder:    defaultBuilder,
	}
}

func (c *ContentTypeClient) Get(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.Do(ctx, http.MethodGet, url, req, opts...)
}

func (c *ContentTypeClient) Post(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.Do(ctx, http.MethodPost, url, req, opts...)
}

func (c *ContentTypeClient) Put(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.Do(ctx, http.MethodPut, url, req, opts...)
}

func (c *ContentTypeClient) PATCH(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.Do(ctx, http.MethodPatch, url, req, opts...)
}

func (c *ContentTypeClient) Delete(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.Do(ctx, http.MethodDelete, url, req, opts...)
}

func (c *ContentTypeClient) Do(ctx context.Context, method, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	return c.do(ctx, method, url, req, opts...)
}

func (c *ContentTypeClient) do(ctx context.Context, method, reqUrl string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	request, bodyBuffer, err := c.parseContentTypeRequest(ctx, method, c.host, reqUrl, req)
	if bodyBuffer != nil {
		defer buffpool.PutBytesBuffer(bodyBuffer)
	}
	if err != nil {
		err = fmt.Errorf("new http request error: %v", err)
		return
	}
	return do(c.rawClient.Client(), request, append(c.requestOptions, opts...)...)
}

func (c *ContentTypeClient) NewRequest(ctx context.Context, method, reqUrl string, req interface{}) (request *http.Request, err error) {
	request, _, err = c.parseContentTypeRequest(ctx, method, c.host, reqUrl, req)
	return
}

func (c *ContentTypeClient) parseContentTypeRequest(ctx context.Context, method, host, url string, req interface{}) (request *http.Request, bodyBuffer *bytes.Buffer, err error) {
	u, err := ihttp.JoinUrl(host, url)
	if err != nil {
		return
	}
	if builder, ok := c.requestBuilderMap[method]; ok {
		return builder(ctx, method, u, req)
	}
	return c.defaultBuilder(ctx, method, u, req)
}
