package httpreq

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/kits/kgo/buffpool"
	"github.com/spelens-gud/Verktyg.git/kits/kurl"
)

var _ ihttp.Client = &JsonClient{}

type JsonClient struct {
	ihttp.Client
}

func NewJsonClient(host string, opts ...ihttp.CliOpt) ihttp.Client {
	return &JsonClient{
		Client: NewContentTypeClient(queryRequestBuilder, map[string]requestBuilder{
			http.MethodPost:  jsonContentParser,
			http.MethodPut:   jsonContentParser,
			http.MethodPatch: jsonContentParser,
		}, host, opts...),
	}
}

func jsonContentParser(ctx context.Context, method string, u *url.URL, req interface{}) (request *http.Request, bodyBuffer *bytes.Buffer, err error) {
	body := parseRawBody(req)
	if body == nil {
		bodyBuffer = buffpool.GetBytesBuffer()
		if err = json.NewEncoder(bodyBuffer).Encode(req); err != nil {
			return
		}
		body = bodyBuffer
	}
	// 按url/query tag拼接query参数  不存在tag则忽略
	query := kurl.Parse2UrlValuesOmitEmptyTag(req, true, "url", "query")
	if request, err = http.NewRequestWithContext(ctx, method, joinUrlQuery(u, query), body); err != nil {
		return
	}
	request.Header.Add(ihttp.HeaderContentType, ihttp.ContentTypeJson)
	return
}
