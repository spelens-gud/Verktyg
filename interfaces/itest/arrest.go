package itest

import (
	"reflect"
	"testing"

	"gotest.tools/assert"
)

var reportList []func()

func AssertErr(t *testing.T, wantError bool, err error) (end bool) {
	t.Helper()
	assert.Assert(t, (!wantError && err == nil) || (wantError && err != nil), err)
	return err != nil
}

func AssertResp(t *testing.T, result TestResult, authFunc interface{}, ret interface{}) {
	t.Helper()
	if err := result.Unmarshal(ret); err != nil {
		t.Fatal(err)
	}
	fv := reflect.ValueOf(authFunc)
	if fv.IsNil() {
		return
	}
	if err, _ := fv.Call([]reflect.Value{reflect.ValueOf(ret).Elem()})[0].Interface().(error); err != nil {
		t.Fatal(err)
	}
}

func ToReport(t *testing.T, result TestResult) {
	t.Helper()
	t.Log(result)
	reportList = append(reportList, result.Report)
}

func Report() {
	for _, f := range reportList {
		f()
	}
}
