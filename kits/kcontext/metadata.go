package kcontext

import (
	"net/http"
	"net/textproto"
	"strings"
)

type (
	Metadata map[string][]string
)

const HeaderRefererService = "X-Referer-Service"

func (data Metadata) GetAll(key string) []string {
	return data[key]
}

func (data Metadata) Get(key string) string {
	return textproto.MIMEHeader(data).Get(key)
}

func (data Metadata) GetJoin(key string, ident string) string {
	return strings.Join(data[key], ident)
}

func (data Metadata) SetHeader(header http.Header, overwrite bool, omitKeys ...string) {
	omitMap := make(map[string]struct{}, len(omitKeys))
	for _, k := range omitKeys {
		omitMap[k] = struct{}{}
	}

	h := textproto.MIMEHeader(data)
	for k := range h {
		v := header.Get(k)
		if _, omit := omitMap[k]; omit || len(v) > 0 && !overwrite {
			continue
		}
		header.Set(k, h.Get(k))
	}
}
