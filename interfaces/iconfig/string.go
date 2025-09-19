package iconfig

import "go.uber.org/atomic"

type configString struct {
	atomic.String
	Init func() string
}

// nolint
func (c *configString) Set(str string) { c.String.Store(str) }

func (c *configString) Get() string {
	ret := c.Load()
	if len(ret) > 0 {
		return ret
	}
	ret = c.Init()
	c.Store(ret)
	return ret
}
