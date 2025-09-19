package go2ts

import (
	"encoding/json"
	"testing"

	"git.bestfulfill.tech/devops/go-core/kits/kdoc/go2ts/testdata"
)

func TestParse(t *testing.T) {
	ret := ParseTsTypes(testdata.HelloParam{})
	b, _ := json.Marshal(ret)
	t.Logf("%s", b)
	re := TsTypes2JsonSchema(ret)
	b, _ = json.Marshal(re)
	t.Logf("%s", b)
}
