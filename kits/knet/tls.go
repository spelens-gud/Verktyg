//go:build !go1.15

package knet

import (
	"context"
	"crypto/tls"
	"net"
)

func dialTls(dialer *net.Dialer, tlsConfig *tls.Config) DialFunc {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return tls.DialWithDialer(dialer, network, addr, tlsConfig)
	}
}
