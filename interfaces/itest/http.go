package itest

import (
	"net/http"

	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
)

type (
	HttpTester interface {
		Test(title, method, url, meta string, param interface{}, opts ...ihttp.ReqOpt) (res TestResult, err error)
	}

	TestResult interface {
		Unmarshal(interface{}) (err error)

		Report()

		Resp() ihttp.Resp
	}

	Reporter interface {
		Report(sample *Sample) (err error)
	}

	Sample struct {
		RawParam   interface{}
		RawRes     interface{}
		QueryParam string
		ResBody    string
		Method     string
		Url        string
		Title      string
		Category   string
		Header     http.Header
	}
)
