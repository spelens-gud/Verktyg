package iyapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/theothertomelliott/acyclic"

	"git.bestfulfill.tech/devops/go-core/interfaces/itest"
	"git.bestfulfill.tech/devops/go-core/kits/kdoc/go2ts"
)

func ParseTestSample2YapiData(st *itest.Sample) (data ApiUploadData) {
	var (
		apiReqData ApiReq
		rb         string
		isJs       bool
		queryTs    = go2ts.ParseTsTypes(st.RawParam)
	)
	spUrls := strings.Split(st.Url, "/")
	for _, u := range spUrls {
		if !strings.HasPrefix(u, ":") {
			continue
		}
		p := strings.TrimPrefix(u, ":")
		if len(queryTs) > 0 {
			for _, f := range queryTs[0].Fields {
				if f.Name == p {
					apiReqData.ReqParams = append(apiReqData.ReqQuery, ReqKVItemSimple{
						Name: p,
						Desc: strings.Join([]string{f.Type, strings.Replace(f.Comment, "//", "", -1)}, "\t"),
					})
					break
				}
			}
		} else {
			apiReqData.ReqParams = append(apiReqData.ReqQuery, ReqKVItemSimple{
				Name: p,
			})
		}
	}

	if st.Method == http.MethodGet {
		if len(queryTs) > 0 {
			for _, f := range queryTs[0].Fields {
				apiReqData.ReqQuery = append(apiReqData.ReqQuery, ReqKVItemSimple{
					Name: f.Name,
					Desc: strings.Join([]string{f.Type, strings.Replace(f.Comment, "//", "", -1)}, "\t"),
				})
			}
		}
	} else {
		qb, _ := json.Marshal(go2ts.TsTypes2JsonSchema(queryTs))
		apiReqData = ApiReq{
			ReqBodyType:         "json",
			ReqBodyIsJsonSchema: true,
			ReqBodyOther:        string(qb),
		}
	}
	for k := range st.Header {
		apiReqData.ReqHeaders = append(apiReqData.ReqHeaders, ReqKVItemDetail{
			ReqKVItemSimple: ReqKVItemSimple{
				Name:    k,
				Value:   k,
				Example: st.Header.Get(k),
				Desc:    k,
			},
			Required: "1",
		})
	}
	js := go2ts.TsTypes2JsonSchema(go2ts.ParseTsTypes(st.RawRes))
	if js != nil && acyclic.Check(js) == nil {
		b, _ := json.Marshal(js)
		rb = string(b)
		isJs = true
	} else {
		rb = st.ResBody
	}
	data = ApiUploadData{
		ApiBase: ApiBase{
			Title:  st.Title,
			Path:   st.Url,
			Method: st.Method,
		},
		ApiReq: apiReqData,
		ApiRes: ApiRes{
			ResBodyType:         "json",
			ResBody:             rb,
			ResBodyIsJsonSchema: isJs,
		},
	}
	return
}
