package httpreq

import (
	"bytes"
	"io"
	"net/url"
	"strings"
)

func joinUrlQuery(u *url.URL, query url.Values) string {
	if len(query) > 0 {
		if len(u.RawQuery) > 0 {
			u.RawQuery += "&"
		}
		u.RawQuery += query.Encode()
	}
	return u.String()
}

func parseRawBody(req interface{}) (body io.Reader) {
	if reader, ok := req.(io.Reader); ok {
		return reader
	}
	switch t := req.(type) {
	case []byte:
		body = bytes.NewReader(t)
	case string:
		body = strings.NewReader(t)
	}
	return
}
