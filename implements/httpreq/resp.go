package httpreq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"git.bestfulfill.tech/devops/go-core/interfaces/ihttp"
	"git.bestfulfill.tech/devops/go-core/kits/kgo"
	"git.bestfulfill.tech/devops/go-core/kits/kgo/buffpool"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

func NewResp(response *http.Response) (res ihttp.Resp, err error) {
	originBody := response.Body
	defer originBody.Close()
	buf := buffpool.GetBytesBuffer()
	if _, err = buf.ReadFrom(originBody); err != nil {
		buffpool.PutBytesBuffer(buf)
		err = RespReadError{
			Response:  response,
			ReadError: err,
		}
		return
	}
	response.Body = buffpool.ReleaseCloser(buf)
	return &Resp{resp: response, bodyBuffer: buf}, nil
}

type RespReadError struct {
	*http.Response
	ReadError error
}

func (resp RespReadError) Error() string {
	return fmt.Sprintf("response body read error: %v", resp.ReadError)
}

type Resp struct {
	resp       *http.Response
	bodyBuffer *bytes.Buffer
}

func (r *Resp) Release() {
	buffpool.PutBytesBuffer(r.bodyBuffer)
}

func (r *Resp) OK() bool {
	return (r.resp.StatusCode / 100) == 2
}

func (r *Resp) Status() int {
	return r.resp.StatusCode
}

func (r *Resp) Body() *bytes.Buffer {
	return r.bodyBuffer
}

func (r *Resp) UnmarshalJson(in interface{}) (err error) {
	if err = json.Unmarshal(r.bodyBuffer.Bytes(), in); err != nil && r.resp.Request != nil {
		logger.FromContext(r.resp.Request.Context()).WithField("body", kgo.UnsafeBytes2string(r.bodyBuffer.Bytes())).
			Error("unmarshal json body error: " + err.Error())
	}
	return
}

func (r *Resp) Header() http.Header {
	return r.resp.Header
}

func (r *Resp) Resp() *http.Response {
	return r.resp
}

var _ ihttp.Resp = &Resp{}
