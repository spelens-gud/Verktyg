package httpreq

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/spelens-gud/Verktyg/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg/kits/kgo/buffpool"
	"github.com/spelens-gud/Verktyg/kits/kurl"
)

var _ ihttp.Client = &FormClient{}

type FormClient struct {
	ihttp.Client
}

func NewFormClient(host string, opts ...ihttp.CliOpt) ihttp.Client {
	return &FormClient{
		Client: NewContentTypeClient(queryRequestBuilder, map[string]requestBuilder{
			http.MethodPost:  formContentParser,
			http.MethodPut:   formContentParser,
			http.MethodPatch: formContentParser,
		}, host, opts...),
	}
}

func formContentParser(ctx context.Context, method string, u *url.URL, req interface{}) (request *http.Request, bodyBuffer *bytes.Buffer, err error) {
	var (
		body  = parseRawBody(req)
		query = kurl.Parse2UrlValuesOmitEmptyTag(req, true, "url", "query")
	)

	if body == nil {
		// form tag拼接表单参数 不存在tag则以字段名为key
		if formValues := kurl.Parse2UrlValues(req, false, "form"); len(formValues) > 0 {
			// 删除query中和form相同的key
			if len(query) > 0 {
				for k := range formValues {
					if _, ok := query[k]; ok {
						query.Del(k)
					}
				}
			}
			bodyBuffer = buffpool.GetBytesBuffer()
			if _, err = bodyBuffer.WriteString(formValues.Encode()); err != nil {
				return
			}
			body = bodyBuffer
		}
	}
	if request, err = http.NewRequestWithContext(ctx, method, joinUrlQuery(u, query), body); err != nil {
		return
	}
	request.Header.Add(ihttp.HeaderContentType, ihttp.ContentTypeForm)
	return
}
