package httptester

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"time"

	"git.bestfulfill.tech/devops/go-core/implements/httpreq"
	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/interfaces/itest"
	"git.bestfulfill.tech/devops/go-core/kits/kurl"
)

var (
	_ itest.HttpTester = &Tester{}
	_ itest.TestResult = &TestResult{}
)

var logger = log.New(os.Stdout, "[HTTPTEST] ", 0)

type (
	// 预请求钩子 根据能够将用户传入的ID并修改头部信息
	RequestHook func(r *http.Request, id string) (err error)
	// 自定义数据包处理 钩子
	DataPacker func(structIn interface{}) (pack interface{})
	// 自定义反序列化函数
	UnmarshalFunc func(in []byte, ret interface{}) (err error)
	// 自定义序列化函数
	MarshalFunc func(ret interface{}) (out []byte, err error)

	Tester struct {
		handler         http.Handler
		requestHook     RequestHook
		dataPacker      DataPacker
		unmarshalFunc   UnmarshalFunc
		marshalFunc     MarshalFunc
		reporter        itest.Reporter
		urlBase         string
		category        string
		defaultID       string
		disableRedirect bool
		timeout         time.Duration
	}

	TestResult struct {
		tester     *Tester
		url        string
		path       string
		method     string
		rawQuery   []byte
		rawRes     []byte
		param      interface{}
		result     interface{}
		mainResult interface{}
		resp       ihttp.Resp
		title      string
		header     http.Header
	}

	Option func(*Tester)
)

func NewHttpTester(baseRoute, category string, opts ...Option) itest.HttpTester {
	ret := &Tester{
		unmarshalFunc: json.Unmarshal,
		marshalFunc:   json.Marshal,
		urlBase:       baseRoute,
		category:      category,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (r *TestResult) String() string {
	var str string
	str += "\n"
	str += fmt.Sprintf("Status: %s\n", r.resp.Resp().Status)
	str += "Header:\n"
	b, _ := json.MarshalIndent(r.resp.Header, "", "\t")
	str += string(b)
	str += "\nBody:\n"
	var tmp interface{}
	_ = json.Unmarshal(r.rawRes, &tmp)
	b, _ = json.MarshalIndent(tmp, "", "\t")
	str += string(b)
	return str
}

func (r *TestResult) Unmarshal(ret interface{}) (err error) {
	packRet := ret
	// 保存结构体准备解析
	if r.tester.dataPacker != nil {
		packRet = r.tester.dataPacker(ret)
	}
	err = r.tester.unmarshalFunc(r.rawRes, packRet)
	if err != nil {
		return
	}
	r.mainResult = packRet
	r.result = ret
	return
}

func (r *TestResult) Resp() ihttp.Resp {
	return r.resp
}

// 上报接口测试详情和结果
func (r *TestResult) Report() {
	sample := itest.Sample{
		RawParam:   r.param,
		RawRes:     r.mainResult,
		QueryParam: string(r.rawQuery),
		ResBody:    string(r.rawRes),
		Method:     r.method,
		Url:        r.path,
		Title:      r.title,
		Category:   r.tester.category,
		Header:     r.header,
	}
	if r.tester.reporter != nil {
		err := r.tester.reporter.Report(&sample)
		if err != nil {
			logger.Printf("上报测试用例文档结果发生错误: %v", err)
			return
		}
	}
}

// 发起http测试 组装结果返回
func (t *Tester) Test(title, method, reqUrl, meta string, param interface{}, opts ...ihttp.ReqOpt) (
	res itest.TestResult, err error) {
	u, err := url.Parse(t.urlBase)
	if err != nil {
		return
	}
	if len(u.Scheme) == 0 {
		u.Scheme = "http"
	}
	if len(u.Host) == 0 {
		u.Host = "localhost"
	}
	u.Path = path.Join(u.Path, reqUrl)
	reqUrl = u.String()
	tr := &TestResult{
		tester: t,
		url:    reqUrl,
		path:   u.Path,
		method: method,
		param:  param,
		title:  title,
	}
	// 处理数据请求
	req, err := tr.parseParam()
	if err != nil {
		return
	}
	// 预请求钩子 提供修改头部
	if t.requestHook != nil {
		err = t.requestHook(req, meta)
		if err != nil {
			return
		}
	}
	for _, o := range opts {
		o(req)
	}

	tr.header = req.Header

	// 发起请求
	ctx := context.Background()
	if t.timeout > 0 {
		var cf func()
		ctx, cf = context.WithTimeout(ctx, t.timeout)
		defer cf()
		req = req.WithContext(ctx)
	}
	if t.handler != nil {
		reqf := func() error {
			recorder := httptest.NewRecorder()
			t.handler.ServeHTTP(recorder, req)
			tr.resp, err = httpreq.NewResp(recorder.Result())
			return err
		}
		err = reqf()
		if err != nil {
			return
		}
		if !t.disableRedirect && ((tr.resp.Status() / 100) == 3) {
			req.URL, _ = url.Parse(tr.resp.Header().Get("Location"))
			err = reqf()
		}
		if err != nil {
			return
		}
	} else {
		tr.resp, err = httpreq.NewRawClient().Do(ctx, req)
		if err != nil {
			return
		}
	}

	// 非200处理
	if tr.resp.Status() != http.StatusOK {
		err = fmt.Errorf("http error:%v\nhedaer:%#v\nbody:\n%s", tr.resp.Status(), tr.resp.Header(), tr.resp.Body().String())
		return
	}
	// 读取返回值
	tr.rawRes = tr.resp.Body().Bytes()
	res = tr
	return
}

// 将参数根据请求种类转为请求body或query
func (r *TestResult) parseParam() (rq *http.Request, err error) {
	switch r.method {
	case http.MethodGet:
		var u *url.URL
		u, err = url.Parse(r.url)
		if err != nil {
			return nil, err
		}
		rawQuery := kurl.Parse2UrlValues(r.param, false, "form", "url").Encode()
		r.rawQuery = []byte(rawQuery)
		u.RawQuery = rawQuery
		rq, err = http.NewRequest(r.method, u.String(), nil)
	default:
		r.rawQuery, err = r.tester.marshalFunc(r.param)
		if err != nil {
			return
		}
		rq, err = http.NewRequest(r.method, r.url, bytes.NewBuffer(r.rawQuery))
	}
	return
}
