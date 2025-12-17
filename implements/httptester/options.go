package httptester

import (
	"net/http"
	"time"

	"github.com/spelens-gud/Verktyg.git/interfaces/itest"
)

func WithDataPacker(d DataPacker) Option {
	return func(tester *Tester) {
		tester.dataPacker = d
	}
}

func WithDefaultID(d string) Option {
	return func(tester *Tester) {
		tester.defaultID = d
	}
}

func WithRequestHook(d RequestHook) Option {
	return func(tester *Tester) {
		tester.requestHook = d
	}
}

func WithReporter(d itest.Reporter) Option {
	return func(tester *Tester) {
		tester.reporter = d
	}
}

func WithUnmarshalFunc(d UnmarshalFunc) Option {
	return func(tester *Tester) {
		tester.unmarshalFunc = d
	}
}
func WithTimeOut(d time.Duration) Option {
	return func(tester *Tester) {
		tester.timeout = d
	}
}

func WithHandler(h http.Handler) Option {
	return func(tester *Tester) {
		tester.handler = h
	}
}

func DisableRedirect() Option {
	return func(tester *Tester) {
		tester.disableRedirect = true
	}
}
