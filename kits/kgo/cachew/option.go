package cachew

import (
	"encoding/json"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type Option struct {
	UnmarshalFunc         func([]byte, interface{}) error   // 反序列化方法
	MarshalFunc           func(interface{}) ([]byte, error) // 序列化方法
	KeyPrefix             string                            // 缓存Key前缀
	AsyncLoad             bool                              // 即使回源过程中 主协程超时也不进行中断 对部分源请求较慢的可以开启
	MaxUpdateCacheRetries int                               // 更新缓存最大重试次数
	DisableUpdateMutex    bool                              // 不使用更新并发检测 对版本时序要求不高的可以开启
	DefaultExpiration     time.Duration                     // 默认过期时间
}

type options []func(option *Option)

var (
	defaultUnmarshalFunc = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal
	defaultMarshalFunc   = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal
)

func defaultOption() *Option {
	return &Option{
		UnmarshalFunc:         defaultUnmarshalFunc,
		MarshalFunc:           defaultMarshalFunc,
		KeyPrefix:             "cache_store:",
		MaxUpdateCacheRetries: 3,
		DefaultExpiration:     time.Minute,
	}
}

func (opts options) Init(opt *Option) *Option {
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithJsoniter(option *Option) {
	option.UnmarshalFunc = jsoniter.Unmarshal
	option.MarshalFunc = jsoniter.Marshal
}

func WithDefaultJson(option *Option) {
	option.UnmarshalFunc = json.Unmarshal
	option.MarshalFunc = json.Marshal
}

func WithFastestJsoniter(option *Option) {
	option.UnmarshalFunc = jsoniter.ConfigFastest.Unmarshal
	option.MarshalFunc = jsoniter.ConfigFastest.Marshal
}
