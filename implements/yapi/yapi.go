package yapi

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/spelens-gud/Verktyg.git/implements/httpreq"
	"github.com/spelens-gud/Verktyg.git/interfaces/ihttp"
	"github.com/spelens-gud/Verktyg.git/interfaces/itest"
	"github.com/spelens-gud/Verktyg.git/interfaces/iyapi"
)

type (
	yapi struct {
		cateStore  map[string]int
		cateLock   sync.Mutex
		httpClient ihttp.Client
		token      string
		projectID  int
		host       string
	}

	Config struct {
		Host      string
		Token     string
		ProjectID int
	}
)

var httpDoOpt = httpreq.WithHttpDoOption(&httpreq.HttpDoOption{
	DisableLog:    true,
	DisableTrace:  true,
	MaxRetryTimes: 1,
})

var o sync.Once

func (cfg Config) NewYapiApi() iyapi.Client {
	return NewYapiApi(cfg.Host, cfg.Token, cfg.ProjectID)
}

func NewYapiApi(host, token string, projectID int) iyapi.Client {
	ret := &yapi{
		host:       host,
		cateStore:  map[string]int{},
		httpClient: httpreq.NewJsonClient(host),
		token:      token,
		projectID:  projectID,
	}
	return ret
}

func (y *yapi) Report(data *itest.Sample) (err error) {
	o.Do(y.InitCate)
	if y.projectID == 0 && len(y.token) == 0 {
		return
	}
	ret := iyapi.ParseTestSample2YapiData(data)
	catID, err := y.GetCateID(data.Category)
	if err != nil {
		return
	}
	ret.CatID = strconv.Itoa(catID)
	return y.UploadYapiApi(&ret)
}

func (y *yapi) InitCate() {
	s := iyapi.CateParam{ProjectID: y.projectID, Token: y.token}
	var ret struct {
		Data []iyapi.CateRet
	}
	resp, err := y.httpClient.Get(context.Background(), iyapi.PostYapiCateUrl, &s, httpDoOpt)
	if err != nil {
		return
	}
	_ = resp.UnmarshalJson(&ret)
	for _, d := range ret.Data {
		y.cateStore[d.Name] = d.Uid
	}
}

func (y *yapi) UploadYapiApi(data *iyapi.ApiUploadData) (err error) {
	data.Token = y.token
	var ret iyapi.Ret
	resp, err := y.httpClient.Post(context.Background(), iyapi.PostYapiUrl, data, httpDoOpt)
	if err != nil {
		return
	}
	err = resp.UnmarshalJson(&resp)
	if ret.Errcode != 0 {
		err = fmt.Errorf("upload failed")
	}
	return
}

func (y *yapi) GetCateID(cate string) (id int, err error) {
	y.cateLock.Lock()
	id, ok := y.cateStore[cate]
	y.cateLock.Unlock()
	if ok {
		return id, nil
	}
	var res iyapi.Ret
	var idRet iyapi.RetID
	res.Data = &idRet
	resp, err := y.httpClient.Post(context.Background(), iyapi.PostYapiAddCateUrl, map[string]interface{}{
		"token":      y.token,
		"name":       cate,
		"project_id": y.projectID,
	}, httpDoOpt)
	if err != nil {
		return
	}
	err = resp.UnmarshalJson(&res)
	if err != nil {
		return
	}
	id = idRet.ID
	if id == 0 {
		err = errors.New("invalid id")
		return
	}
	y.cateLock.Lock()
	y.cateStore[cate] = id
	y.cateLock.Unlock()
	return
}
