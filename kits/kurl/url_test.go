package kurl

import (
	"testing"
)

type AnonymousField2 struct {
	Int2 int `url:"int2"`
}

type FF string

type AnonymousField struct {
	FF
	AnonymousField2
	unExportInt2 int
	StrJsonTag2  string `json:"strJsonTag2"`
	StrFormTag2  string `form:"strFormTag2"`
	StrUrlTag2   string `url:"strUrlTag2"`
	StrUrlTag3   string `url:"strUrlTag3,omitempty"`
}

func TestParseUrl2(t *testing.T) {
	t.Log(Parse2UrlValues(map[string]string{
		"test": "Test",
	}, true).Encode())
	t.Log(Parse2UrlValues(map[string]string{
		"test": "Test",
	}, false).Encode())
	t.Log(Parse2UrlValues(map[string]int{
		"test": 215251,
	}, false).Encode())
	t.Log(Parse2UrlValues(map[string]float64{
		"test": 215251,
	}, false).Encode())
	t.Log(Parse2UrlValues(map[string]bool{
		"test": true,
	}, false).Encode())
}

func TestParseUrl3(t *testing.T) {
	t.Log(Parse2UrlValuesOmitEmptyTag(AnonymousField{
		AnonymousField2: AnonymousField2{},
		unExportInt2:    0,
		StrJsonTag2:     "",
		StrFormTag2:     "",
		StrUrlTag3:      "ff",
		StrUrlTag2:      "",
	}, false, "url", "query").Encode())
}

func TestParseUrl(t *testing.T) {
	in := struct {
		Str         string
		unExportStr string
		Int         int
		unExportInt int
		OmitInt     int    `url:"-"`
		StrJsonTag  string `json:"strJsonTag"`
		StrFormTag  string `form:"strFormTag"`
		StrUrlTag   string `url:"strUrlTag"`
		AnonymousField
		Struct AnonymousField `url:"struct"`
		Slice  []AnonymousField
	}{AnonymousField: AnonymousField{unExportInt2: 0}}
	ret := Parse2UrlValues(in, false, "url", "form").Encode()
	ret2 := ParseStruct2UrlValues(in, false, "url", "form").Encode()
	if ret != ret2 {
		t.Fatal("")
	}
	t.Log(ret)
	t.Log(ret2)
	t.Log(ParseStruct2UrlValues(in, true, "url", "form").Encode())
}
