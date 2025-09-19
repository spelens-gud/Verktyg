package kjson

import (
	"fmt"
	"testing"
)

func TestRefLoad(t *testing.T) {
	ret, err := LoadRefJson([]byte(
		//language=json
		`{"test":"${testkey}","ffff":"sagsda","fff":{"$ref":"testkey2","normal":"fasdasdadsg","fff": "sss"}}`),
		func(string string) (ref []byte, err error) {
			return []byte(fmt.Sprintf(`{"fff":"%s","ddd":"tttt"}`, string)), nil
		})
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	t.Logf("%s", ret)

}
