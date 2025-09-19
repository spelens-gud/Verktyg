package kserver

import (
	"net"
	"sync"
)

type Listener struct {
	net.Listener
	once sync.Once

	accepted chan struct{}
}

func NewAcceptedNoticeListener(listener net.Listener) *Listener {
	l := &Listener{
		Listener: listener,
		accepted: make(chan struct{}),
	}
	return l
}

func (l *Listener) Accepted() <-chan struct{} {
	return l.accepted
}

func (l *Listener) Accept() (net.Conn, error) {
	l.once.Do(func() {
		close(l.accepted)
	})
	return l.Listener.Accept()
}
