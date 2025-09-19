# Http请求库

## 请求使用

```go
	// 新建一个client form client用于提交form格式client
	formClient := httpreq.NewFormClient("http://localhost")
	// jsonClient
	jsonClient := httpreq.NewJsonClient("http://localhost")

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
	// 第三个参数 接受结构体 或 map 会自动解析并解析为form/json 放到body 识别tag form/json
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
```

## url body自动拼装逻辑

GET 请求 tag url/form/query 都会被添加到query参数

```go
	resp, _ := cli.Get(context.Background(), "/path?query=a", struct {
		A string `url:"a"`
		B string `form:"b"`
		C string `query:"c"`
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a&a=&b=&c=
```

GET 请求 无tag 会使用字段名添加到query参数

```go
	resp, _ = cli.Get(context.Background(), "/path?query=a", struct {
		A string `url:"a"`
		B string ``
		C string `query:"c"`
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a&B=&a=&c=
```

GET 请求 将tag置为 - 则排除query参数

```go
	resp, _ = cli.Get(context.Background(), "/path?query=a", struct {
		A string `url:"a"`
		B string `url:"-"`
		C string `query:"c"`
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a&a=&c=
```


POST请求 tag url/query 会被添加到query参数

tag form会被放到body 无form tag 会默认使用字段名
	
```go
	resp, _ = cli.Post(context.Background(), "/path?query=a", struct {
		A string `url:"a"`
		B string `form:"b"`
		C string `query:"c"`
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a&a=&c=
	t.Log(resp.Body().String())             // A=&C=&b=
```

POST请求 无url/query tag 不会将字段名放到query参数

tag form会被放到body 无form tag 会默认使用字段名

```go
	resp, _ = cli.Post(context.Background(), "/path?query=a", struct {
		A string
		B string `form:"b"`
		C string
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a
	t.Log(resp.Body().String())             // A=&C=&b=
```

POST请求 使用 - 过滤body参数

```go
	resp, _ = cli.Post(context.Background(), "/path?query=a", struct {
		A string `form:"-"`
		B string `form:"b"`
		C string
	}{})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a
	t.Log(resp.Body().String())             // C=&b=
```

POST请求 使用 omitempty 过滤空参数

```go
	resp, _ = cli.Post(context.Background(), "/path?query=a", struct {
		A string `url:"a,omitempty"`
		B string `form:"b"`
		C string `form:"c,omitempty"`
	}{
		A: "a",
	})
	t.Log(resp.Resp().Request.URL.String()) // http://localhost/path?query=a
	t.Log(resp.Body().String())             // a=a&b=
```

http 断路器使用
说明：
1. 断路器关闭状态，请求正常发出，错误数在timeout时间内满足条件(errorThreshold)，断路器被设置为打开状态（错误数，成功数都重置为0）
2. 断路器打开状态，，所有请求都中断，睡眠timeout 时间后，断路器会处于半开关状态（错误数，成功数都重置为0）
3. 断路器半打开关状态，请求正常发出，在timeout时间内连续成功的数量达到条件(successThreshold)，就能关闭断路器，中间有一次失败断路器会设置成打开状态

例子1：
```go
    type client struct {
        breaker httpreq.HttpBreaker
        cli ihttp.Client
    }
	
    cli := client{
        cli:     httpreq.NewRawClient(o.CliOption...),
        breaker: httpreq.NewHttpBreaker(o.BreakerOption...),
    }
	
    func TestBreaker() {
        err = c.breaker.Run(func() (err error) {
            resp, err := c.cli.Get(ctx, path, data)
            if err != nil {
                logger.FromContext(ctx).WithTag("CONFIG_REQ_ERR").Errorf("get config error: %v", err)
                return
            }
            ct := new(Content)
            if err = c.parseResp(resp, ct); err != nil {
                return
            }
            return
        })
        return
    }
```
例子2：
```go
    type client struct {
        cli ihttp.Client
    }
	
    cli := client{
		cli:     httpreq.NewBreakerJsonClient(
			httpreq.NewHttpBreaker(o.BreakerOption...),
			endpoint,o.CliOption...),
    }
	
    func TestBreaker(ctx context.Context, url string, req interface{}, opts ...ihttp.ReqOpt) (resp ihttp.Resp,err error) {
        return c.cli.Get(ctx, url, req, opts...)
    }
```