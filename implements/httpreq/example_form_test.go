package httpreq

import (
	"context"
	"testing"
)

func TestForm(t *testing.T) {
	// 新建一个client form client用于提交form格式client
	formClient := NewFormClient("http://localhost")

	// 可以通过Raw调用和go原生接近的client
	rawClient := formClient.Raw()

	// 可以再通过Client获取真实的原生http.Client
	_ = rawClient.Client()

	param := struct {
		Param2 string `url:"param2"`
	}{
		Param2: "xxxx",
	}
	// get方法
	// 第一个参数 接受context.Context
	// 第二个参数 接受字符串url  可以补充url路径 也可以添加一些query参数
	// 第三个参数 接受结构体 或 map 会自动解析并添加到url的query tag 优先级   url > query > form
	_, err := formClient.Get(context.Background(), "/xxx?param=xxx", param)
	if err != nil {
		t.Fatal(err)
	}
	// post等带body方法
	// 第三个参数 接受结构体 或 map 会自动解析并解析为form 放到body  tag优先级 form > url
	resp, err := formClient.Post(context.Background(), "/xxx?param=xxx", param)
	if err != nil {
		t.Fatal(err)
	}
	// 判断返回码是否 2XX
	if !resp.OK() {
		t.Fatal(resp.Body().String())
	}
	// 打印body
	t.Log(resp.Body().String())
	// json解析  注意 json解析后body会为空 如有需要复用body 请提前复制
	err = resp.UnmarshalJson(&struct {
	}{})
	if err != nil {
		t.Fatal(err)
	}
}
