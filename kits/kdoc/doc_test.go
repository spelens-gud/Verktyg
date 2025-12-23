package kdoc

import (
	"testing"

	"github.com/spelens-gud/Verktyg/kits/kdoc/testdata"
)

func TestGetStructFieldsDoc(t *testing.T) {
	res := GetStructFieldsDoc(&testdata.Test{})
	t.Logf("%+v", res)
}
