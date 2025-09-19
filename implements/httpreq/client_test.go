package httpreq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type s struct {
}

func (s s) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_ = request.Body.Close()
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write([]byte("sasas"))
}

var _ http.Handler = s{}

func TestBytes(t *testing.T) {
	b := new(bytes.Buffer)
	b.WriteString("test")

	bys := b.Bytes()

	t.Logf("%s", bys)
	t.Logf("%s", b)
	_, _ = ioutil.ReadAll(b)
	t.Logf("%s", bys)
	t.Logf("%s", b)
}

func TestClientError(t *testing.T) {
	cli := NewRawClient()
	ctx, cf := context.WithCancel(context.Background())
	cf()
	_, err := cli.Get(ctx, "http://xxx.com")
	if err != nil {
		if err.(*url.Error).Err != context.Canceled {
			t.Error(err)
		}
	}
}

func TestGetBackoffTime(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Log(getBackoffTime(i, 504).String())
	}
}

func TestTransport(t *testing.T) {
	logger.SetLevel(ilog.Info)
	cli := NewRawClient()
	req, _ := http.NewRequest("GET", "http://www.baidu.com", nil)
	resp, err := RoundTrip(cli.Client().Do, req)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDoTransportOpt(t *testing.T) {
	cli := NewRawClient()
	req, _ := http.NewRequest("GET", "http://www.baidu.com", nil)
	resp, err := DoRequest(cli.Client(), req, WithHttpDoOption(&HttpDoOption{
		DisableTrace:      true,
		DisableMetaHeader: true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, cf := context.WithTimeout(ctx, time.Second)
	defer cf()
	cli := NewFormClient("http://scm.37api.com.cn?test=xxx")
	resp, err := cli.Get(ctx, "/api/v1/config/origin?xzx=23532", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Release()
	t.Fatal(resp.Body())
}

func TestProxy(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", s{})
	}()

	cli := NewRawClient()

	_, err := cli.Get(context.Background(), "http://www.baidu.com/")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := cli.Get(context.Background(), "http://www.baidu.com/")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", resp.Header().Get(ihttp.HeaderContentLength))
	//t.Logf(resp.Body().String())
}

func TestRetry(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", s{})
	}()

	cli := NewRawClient()

	resp, err := cli.Get(context.Background(), "http://localhost/xxx?sasa=zxvzxz#saassa")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", resp.Header().Get("Content-Length"))
}

func TestRetryPost(t *testing.T) {
	go func() {
		e := gin.New()
		e.POST("/r", func(c *gin.Context) {
			b, _ := ioutil.ReadAll(c.Request.Body)
			t.Logf("%s", b)
			c.AbortWithStatus(500)
		})
		e.POST("", func(c *gin.Context) {
			t.Logf("redirect")
			c.Redirect(307, "/r")
		})
		_ = e.Run(":80")
	}()

	cli := NewJsonClient("http://localhost", ihttp.WithTransportSetting(func(transport *http.Transport) {
	}))
	te := struct {
		Test string
	}{"xx"}

	bf := new(bytes.Buffer)
	_ = json.NewEncoder(bf).Encode(&te)

	ctx := context.Background()
	resp, err := cli.Post(ctx, "", ioutil.NopCloser(bf), WithHttpDoOption(&HttpDoOption{
		MaxRetryTimes: 10,
	}))
	if err != nil {

		t.Fatal(err)
	}
	t.Logf("%v", resp.Header().Get("Content-Length"))
}

func TestRequest(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", s{})
	}()

	cli := NewFormClient("http://forum.com.cn/api/attachments?uid=1806136521", func(r *http.Client) {
		r.Timeout = 10 * time.Second
	})

	resp, err := cli.Post(context.Background(), "", struct {
		Type int
	}{
		Type: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestRequestError(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", s{})
	}()

	cli := NewFormClient("http://1.2.3.4", func(cli *http.Client) {
		cli.Timeout = time.Second
	})

	resp, err := cli.Post(context.Background(), "", struct {
		Type int
	}{
		Type: 1,
	})
	if err != nil {
		return
	}
	t.Fatal(resp)
}

func TestRedirect(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Location", "http://localhost")
			writer.WriteHeader(302)
		}))
	}()

	cli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return ihttp.NewRedirectError("xxx")
		},
	}

	req, _ := http.NewRequest("GET", "http://localhost/", nil)
	_, err := doRequest(cli, req)
	if err != nil {
		t.Log(err)
	}
}

func TestRespJson(t *testing.T) {
	go func() {
		_ = http.ListenAndServe(":80", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			_, _ = writer.Write([]byte("invalid json"))
			writer.WriteHeader(200)
		}))
	}()

	cli := &http.Client{}
	req, _ := http.NewRequest("GET", "http://127.0.0.1", nil)
	resp, err := do(cli, req)
	if err != nil {
		t.Fatal(err)
	}
	var ret struct {
		Data string `json:"data"`
	}

	t.Log(resp.Body().String())
	err = resp.UnmarshalJson(&ret)
	if err != nil {
		t.Log(err)
	}
	t.Log(resp.Body().String())
}

func TestBreaker(t *testing.T) {
	c := NewBreakerJsonClient(NewHttpBreaker(), "http://127.0.0.1:108/")
	for i := 0; i < 20; i++ {
		_, err := c.Get(context.Background(), "/api/job-executor/v1/process/cmdline?pid=377", nil)
		fmt.Println(err)
	}
}
