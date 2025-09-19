package kdoc

import (
	"testing"

	"git.bestfulfill.tech/devops/go-core/kits/kdoc/testdata"
)

func TestGetStructFieldsDoc(t *testing.T) {
	res := GetStructFieldsDoc(&testdata.Test{})
	t.Logf("%+v", res)
}
