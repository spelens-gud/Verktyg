package itrace

import (
	"net/http"
	"testing"
)

func TestGetRequestIDFromMetadata(t *testing.T) {
	header := map[string][]string{}
	http.Header(header).Set(HeaderXRequestID, "ddd")
	if GetRequestIDFromMetadata(header) != "ddd" {
		t.Fatal()
	}
	http.Header(header).Set("abc", "xxx")
	if GetRequestIDFromMetadata(header, "abc") != "xxx" {
		t.Fatal()
	}

}
