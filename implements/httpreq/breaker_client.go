package httpreq

import (
	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
)

var _ ihttp.Client = &BreakerJsonClient{}

type BreakerJsonClient struct {
	ihttp.Client
}

func NewBreakerJsonClient(breaker HttpBreaker, host string, opts ...ihttp.CliOpt) ihttp.Client {
	return &BreakerJsonClient{
		Client: NewBreakerContentTypeClient(NewJsonClient(host, opts...), breaker),
	}
}

var _ ihttp.Client = &BreakerFormClient{}

type BreakerFormClient struct {
	ihttp.Client
}

func NewBreakerFormClient(breaker HttpBreaker, host string, opts ...ihttp.CliOpt) ihttp.Client {
	return &BreakerJsonClient{
		Client: NewBreakerContentTypeClient(NewFormClient(host, opts...), breaker),
	}
}
