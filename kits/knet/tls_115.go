//go:build go1.15

package knet

import (
	"crypto/tls"
	"net"
)

func dialTls(dialer *net.Dialer, tlsConfig *tls.Config) DialFunc {
	tlsDialer := &tls.Dialer{
		NetDialer: dialer,
		Config:    tlsConfig,
	}
	return tlsDialer.DialContext
}
