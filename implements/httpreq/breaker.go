package httpreq

import (
	"context"
	"net/url"
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/ierror"

	"github.com/eapache/go-resiliency/breaker"

	"github.com/spelens-gud/Verktyg.git/kits/kerror/errorx"
)

type BreakerOption struct {
	Timeout          time.Duration
	SuccessThreshold int
	FailedThreshold  int
	BreakErrors      []func(error) bool
}

type HttpBreaker interface {
	Run(func() error, ...func(option *ReqOption)) error
	Go(func() error, ...func(option *ReqOption)) error
}

type httpBreaker struct {
	b   *breaker.Breaker
	opt *BreakerOption
}

type ReqOption struct {
	LogBreak func()
}

func WithLogBreak(logBreak func()) func(option *ReqOption) {
	return func(option *ReqOption) {
		option.LogBreak = logBreak
	}
}

func (h *httpBreaker) Go(f func() error, ops ...func(option *ReqOption)) error {
	opt := ReqOption{}
	for _, op := range ops {
		op(&opt)
	}
	return h.exec(h.b.Go, f, opt)
}

func (h *httpBreaker) Run(f func() error, ops ...func(option *ReqOption)) error {
	opt := ReqOption{}
	for _, op := range ops {
		op(&opt)
	}
	return h.exec(h.b.Run, f, opt)
}

func (h *httpBreaker) exec(executor func(func() error) error, work func() error, opt ReqOption) error {
	var execErr error
	breakerErr := executor(func() error {
		err := work()
		execErr = err

		if urlErr, is := err.(*url.Error); is && urlErr.Err != context.Canceled {
			return err
		}

		for _, breakError := range h.opt.BreakErrors {
			if breakError(err) {
				return err
			}
		}
		return nil
	})
	switch breakerErr {
	case breaker.ErrBreakerOpen:
		if opt.LogBreak != nil {
			opt.LogBreak()
		}
		return errorx.Wrap(breakerErr, ierror.ResourceExhausted)
	default:
		return execErr
	}
}

func IsBreakOpenErr(err error) bool {
	if ex, ok := err.(ierror.CodeError); ok && ex != nil {
		return ex.Unwrap() == breaker.ErrBreakerOpen
	}
	return false
}

func NewHttpBreaker(opts ...func(option *BreakerOption)) HttpBreaker {
	o := &BreakerOption{
		Timeout:          time.Second * 10,
		SuccessThreshold: 1,
		FailedThreshold:  10,
	}
	for _, opt := range opts {
		opt(o)
	}
	ret := &httpBreaker{
		opt: o,
	}
	ret.b = breaker.New(o.FailedThreshold, o.SuccessThreshold, o.Timeout)
	return ret
}
