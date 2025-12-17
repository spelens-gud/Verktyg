package iyapi

import (
	"github.com/spelens-gud/Verktyg.git/interfaces/itest"
)

const (
	PostYapiUrl        = "/api/interface/save"
	PostYapiCateUrl    = "/api/interface/getCatMenu"
	PostYapiAddCateUrl = "/api/interface/add_cat"
)

type (
	CateParam struct {
		ProjectID int    `url:"project_id"`
		Token     string `url:"token"`
	}

	CateRet struct {
		Name string `json:"name"`
		Uid  int    `json:"_id"`
	}

	Ret struct {
		Errcode int         `json:"errcode"`
		Data    interface{} `json:"data"`
	}

	RetID struct {
		ID int `json:"_id"`
	}

	Client interface {
		Report(data *itest.Sample) (err error)

		UploadYapiApi(data *ApiUploadData) (err error)

		GetCateID(cate string) (id int, err error)
	}
)
