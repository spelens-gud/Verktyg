package httpreq

import (
	"context"
	"net/http"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"

	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
)

var _ ihttp.Client = &BreakerContentTypeClient{}

type BreakerContentTypeClient struct {
	client  ihttp.Client
	breaker HttpBreaker
}

func NewBreakerContentTypeClient(client ihttp.Client, breaker HttpBreaker) ihttp.Client {
	return &BreakerContentTypeClient{
		client:  client,
		breaker: breaker,
	}
}

func (c BreakerContentTypeClient) Raw() ihttp.RawClient {
	return c.client.Raw()
}

func (c BreakerContentTypeClient) Get(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.Get(ctx, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], uri[%s]", http.MethodGet, url)
	}))
	return
}

func (c BreakerContentTypeClient) Post(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.Post(ctx, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], uri[%s]", http.MethodPost, url)
	}))
	return
}

func (c BreakerContentTypeClient) Put(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.Put(ctx, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], uri[%s]", http.MethodPut, url)
	}))
	return
}

func (c BreakerContentTypeClient) PATCH(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.PATCH(ctx, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], uri[%s]", http.MethodPatch, url)
	}))
	return
}

func (c BreakerContentTypeClient) Delete(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.Delete(ctx, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], url[%s]", http.MethodDelete, url)
	}))
	return
}

func (c BreakerContentTypeClient) Do(ctx context.Context, method, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp, err error) {
	err = c.breaker.Run(func() (err error) {
		resp, err = c.client.Do(ctx, method, url, req, opts...)
		return
	}, WithLogBreak(func() {
		logger.FromContext(ctx).WithTag(ilog.TagCircuitBreaker).Errorf("req circuit breaker, method[%s], url[%s]", method, url)
	}))
	return
}

func (c BreakerContentTypeClient) WithRequestOptions(opts ...ihttp.ReqOpt) ihttp.Client {
	return c.client.WithRequestOptions(opts...)
}

func (c BreakerContentTypeClient) NewRequest(ctx context.Context, method, reqUrl string, req interface{}) (request *http.Request, err error) {
	return c.client.NewRequest(ctx, method, reqUrl, req)
}
