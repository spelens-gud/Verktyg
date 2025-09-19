package main

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"git.bestfulfill.tech/devops/go-core/implements/httpreq"
	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var (
	port = strconv.Itoa(rand.Intn(10000) + 10000)
	addr = "localhost:" + port

	ctx = context.Background()
)

func init() {
	go func() {
		_ = http.ListenAndServe(addr, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// 校验头部
			if request.Header.Get("Auth") != "secret" {
				writer.WriteHeader(500)
				return
			}
			writer.WriteHeader(200)
			_, _ = writer.Write([]byte(`success`))
		}))
	}()
}

func main() {
	client := httpreq.NewJsonClient("http://"+addr).WithRequestOptions(
		// 使用固定的请求选项
		httpreq.WithHttpDoOption(&httpreq.HttpDoOption{
			DisableLog: true,
		}), func(r *http.Request) {
			// 头部固定秘钥
			r.Header.Set("Auth", "secret")
		},
	)

	// 请求
	resp, err := client.Get(ctx, "/test", nil)
	if err != nil {
		panic(err)
	}
	// resp: [ 200 ] success
	logger.FromContext(ctx).Infof("resp: [ %v ] %s", resp.Status(), resp.Body().String())

	// 原生客户端
	rawClient := &http.Client{}
	// 通过RoundTripper注入监控日志链路追踪自动重试
	rawClient.Transport = ihttp.WrapRoundTripper(rawClient.Transport, httpreq.RoundTrip)

	// 请求
	rawResp, err := rawClient.Get("http://" + addr)
	if err != nil {
		panic(err)
	}
	defer rawResp.Body.Close()

	// 读取返回
	data, err := ioutil.ReadAll(rawResp.Body)
	if err != nil {
		panic(err)
	}

	// resp: [ 500 ]
	logger.FromContext(ctx).Infof("resp: [ %v ] %s", rawResp.StatusCode, data)

}
