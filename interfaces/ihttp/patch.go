package ihttp

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

func init() {
	http.DefaultClient = NewDefaultHttpClient()
}

func WithProxy(proxy func(request *http.Request) (*url.URL, error)) CliOpt {
	return WithTransportSetting(func(transport *http.Transport) {
		transport.Proxy = proxy
	})
}

func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) CliOpt {
	return WithTransportSetting(func(transport *http.Transport) {
		transport.DialContext = dialer
	})
}

func WithTransportSetting(transportPatch func(*http.Transport)) CliOpt {
	return func(cli *http.Client) {
		if t, ok := UnwrapTransport(cli.Transport); ok {
			transportPatch(t)
		}
	}
}
