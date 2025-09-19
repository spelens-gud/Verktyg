package incontext

import "net/http"

func (key Key) WithHttpRequest(req *http.Request, value interface{}) *http.Request {
	return req.WithContext(key.WithValue(req.Context(), value))
}

func (key Key) FromHttpRequest(req *http.Request) interface{} {
	return key.Value(req.Context())
}
