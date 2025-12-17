package knet

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/spelens-gud/Verktyg/kits/klog/logger"
)

type DialFunc func(ctx context.Context, network, addr string) (net.Conn, error)

func (f DialFunc) Dial(network, addr string) (net.Conn, error) {
	return f(logger.AddCallerSkip(context.Background(), 1), network, addr)
}

func (f DialFunc) String() string {
	return "sy-core:knet.WrapTcpDialer"
}

var (
	mu       sync.Mutex
	wrapFunc []func(DialFunc) DialFunc
)

func RegisterDialerWrapFunc(fs ...func(DialFunc) DialFunc) {
	mu.Lock()
	wrapFunc = append(wrapFunc, fs...)
	mu.Unlock()
}

func WrapTcpDialer(timeout, keepAlive time.Duration, tlsConfig *tls.Config) DialFunc {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: keepAlive,
	}
	dialFunc := dialer.DialContext

	if tlsConfig != nil {
		dialFunc = dialTls(dialer, tlsConfig)
	}

	mu.Lock()
	for _, w := range wrapFunc {
		dialFunc = w(dialFunc)
	}
	mu.Unlock()

	return func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		conn, err = dialFunc(ctx, network, addr)
		if err != nil {
			logger.FromContext(ctx).WithTag("NET_ERR").AddCallerSkip(1).Error(err.Error())
		}
		return
	}
}
