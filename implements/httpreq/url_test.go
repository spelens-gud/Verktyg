package httpreq

import (
	"context"
	"io"
	"net/http"
	"testing"

	"gotest.tools/assert"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var cli = NewFormClient("http://localhost")

var h = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
	b, _ := io.ReadAll(request.Body)
	_, _ = writer.Write(b)
})

func TestGetUrl(t *testing.T) {
	logger.SetLevel(ilog.Error)
	go func() {
		_ = http.ListenAndServe(":80", h)
	}()
	t.Run("GET 请求 tag url/form/query 都会被添加到query参数", func(t *testing.T) {
		resp, _ := cli.Get(context.Background(), "/path?query=a", struct {
			A string `url:"a"`
			B string `form:"b"`
			C string `query:"c"`
		}{})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&a=&b=&c=")
	})

	t.Run("GET 请求 无tag 会使用字段名添加到query参数", func(t *testing.T) {
		resp, _ := cli.Get(context.Background(), "/path?query=a", struct {
			A string `url:"a"`
			B string ``
			C string `query:"c"`
		}{})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&B=&a=&c=")
	})

	t.Run("GET 请求 将tag置为 - 则排除query参数", func(t *testing.T) {
		resp, _ := cli.Get(context.Background(), "/path?query=a", struct {
			A string `url:"a"`
			B string `url:"-"`
			C string `query:"c"`
		}{})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&a=&c=")
	})

	t.Run("POST请求 tag url/query 会被添加到query参数", func(t *testing.T) {
		// tag form会被放到body 无form tag 会默认使用字段名
		resp, _ := cli.Post(context.Background(), "/path?query=a", struct {
			A string `url:"a"`
			B string `form:"b"`
			C string `query:"c"`
		}{})
		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&a=&c=")
		assert.Equal(t, resp.Body().String(), "A=&C=&b=")
	})

	t.Run("POST请求 无url/query tag 不会将字段名放到query参数", func(t *testing.T) {
		// tag form会被放到body 无form tag 会默认使用字段名
		resp, _ := cli.Post(context.Background(), "/path?query=a", struct {
			A string
			B string `form:"b"`
			C string
		}{})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a")
		assert.Equal(t, resp.Body().String(), "A=&C=&b=")
	})

	t.Run("POST请求 使用 - 过滤body参数", func(t *testing.T) {
		resp, _ := cli.Post(context.Background(), "/path?query=a", struct {
			A string `form:"-"`
			B string `form:"b"`
			C string
		}{})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a")
		assert.Equal(t, resp.Body().String(), "C=&b=")
	})

	t.Run("POST请求 使用 omitempty 过滤空参数", func(t *testing.T) {
		resp, _ := cli.Post(context.Background(), "/path?query=a", struct {
			A string `url:"a,omitempty" form:"-"`
			B string `form:"b"`
			C string `form:"c,omitempty"`
		}{
			A: "a",
		})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&a=a")
		assert.Equal(t, resp.Body().String(), "b=")
	})

	t.Run("POST请求 使用 map", func(t *testing.T) {
		resp, _ := cli.Post(context.Background(), "/path?query=a", map[string]string{
			"a":     "a",
			"b":     "b",
			"query": "c",
		})

		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a")
		assert.Equal(t, resp.Body().String(), "a=a&b=b&query=c")
	})

	t.Run("GET请求 使用 map", func(t *testing.T) {
		resp, _ := cli.Get(context.Background(), "/path?query=a", map[string]string{
			"a":     "a",
			"b":     "b",
			"query": "c",
		})
		assert.Equal(t, resp.Resp().Request.URL.String(), "http://localhost/path?query=a&a=a&b=b&query=c")
	})
}
