package iyapi

type ReqKVItemSimple struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Example string `json:"example"`
	Desc    string `json:"desc"`
}

type ReqKVItemDetail struct {
	ReqKVItemSimple
	Type     string `json:"type"`
	Required string `json:"required"`
}

type ApiBase struct {
	ID        int      `json:"_id,omitempty" `
	UID       int      `json:"uid,omitempty"`
	CatID     string   `json:"catid,omitempty"`
	ProjectID int      `json:"project_id,omitempty"`
	EditUID   int      `json:"edit_uid,omitempty"`
	AddTime   int      `json:"add_time,omitempty"`
	UpTime    int      `json:"up_time,omitempty"`
	Status    string   `json:"status,omitempty"`
	Title     string   `json:"title,omitempty"`
	Path      string   `json:"path,omitempty"`
	Method    string   `json:"method,omitempty"`
	Tag       []string `json:"tag,omitempty"`
	Desc      string   `json:"desc,omitempty"`
}

type ApiReq struct {
	ReqParams           []ReqKVItemSimple `json:"req_params,omitempty"`
	ReqHeaders          []ReqKVItemDetail `json:"req_headers,omitempty"`
	ReqQuery            []ReqKVItemSimple `json:"req_query,omitempty"`
	ReqBodyForm         []ReqKVItemDetail `json:"req_body_form,omitempty"`
	ReqBodyIsJsonSchema bool              `json:"req_body_is_json_schema,omitempty"`
	ReqBodyType         string            `json:"req_body_type,omitempty"`
	ReqBodyOther        string            `json:"req_body_other,omitempty"`
}

type ApiRes struct {
	ResBodyIsJsonSchema bool   `json:"res_body_is_json_schema,omitempty"`
	ResBodyType         string `json:"res_body_type,omitempty"`
	ResBody             string `json:"res_body,omitempty"`
}

type ApiUploadData struct {
	Token string `json:"token"`
	ApiBase
	ApiReq
	ApiRes
}

type ApiRet struct {
	ErrCode int           `json:"errcode"`
	ErrMsg  string        `json:"errmsg"`
	Data    ApiUploadData `json:"data"`
}

type InterfaceParam struct {
	Token string `url:"token"`
	ID    int    `url:"id"`
}

type InterfaceListData struct {
	Count int             `json:"count"`
	Total int             `json:"total"`
	List  []ApiUploadData `json:"list,omitempty"`
}

type InterfaceList struct {
	ErrCode int               `json:"errcode"`
	ErrMsg  string            `json:"errmsg"`
	Data    InterfaceListData `json:"data"`
}

type InterfaceListParam struct {
	Token string `url:"token,omitempty"`
	CatID int    `url:"catid,omitempty"`
	Page  int    `url:"Page"`
	Limit int    `url:"limit"`
}
