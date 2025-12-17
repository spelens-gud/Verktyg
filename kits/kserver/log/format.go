package log

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
)

const logTag = "HTTP_SERVER"

type ServerLoggerInfo struct {
	StatusCode     int     `json:"status"`
	Time           string  `json:"time"`
	ClientIP       string  `json:"client_ip"`
	RemoteAddr     string  `json:"remote_addr"`
	Method         string  `json:"method"`
	RequestID      string  `json:"request_id"`
	RemoteHost     string  `json:"remote_host"`
	Referer        string  `json:"referer,omitempty"`
	RefererService string  `json:"referer_service,omitempty"`
	Cost           float64 `json:"cost"`
	Duration       string  `json:"duration"`
	Agent          string  `json:"agent"`
	Bytes          int64   `json:"bytes"`
	Version        string  `json:"http_version"`
	Uri            string  `json:"uri"`
	Request        string  `json:"request"`
	PostData       string  `json:"post_data,omitempty"`
	CookieData     string  `json:"cookie_data,omitempty"`
	Error          string  `json:"error,omitempty"`

	Tag         string `json:"tag"`
	Application string `json:"application"`
}

func (info *ServerLoggerInfo) Marshal() []byte {
	info.Tag = logTag
	info.Application = iconfig.GetApplicationName()
	str, _ := jsoniter.ConfigFastest.Marshal(info)
	return str
}

func (info *ServerLoggerInfo) Reset() {
	*info = ServerLoggerInfo{}
}
