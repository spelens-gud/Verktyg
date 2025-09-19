package kjson

import (
	"strings"

	jsoniter "github.com/json-iterator/go"
)

const (
	refKey = "$ref"
	prefix = "${"
	suffix = "}"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type RefLoadFunc func(string string) (ref []byte, err error)

func LoadRefJson(in []byte, loadFunc RefLoadFunc) (out []byte, err error) {
	var rootMap map[string]interface{}
	err = json.Unmarshal(in, &rootMap)
	if err != nil {
		return
	}
	for k, value := range rootMap {
		switch v := value.(type) {
		// json
		case map[string]interface{}:
			ref, ok := v[refKey].(string)
			if !ok {
				continue
			}
			// 获取关联数据
			refData, err := loadFunc(ref)
			if err != nil {
				return nil, err
			}
			// 将关联数据嵌入到当前map
			var refSt map[string]interface{}
			err = json.Unmarshal(refData, &refSt)
			if err != nil {
				return nil, err
			}
			for refK, refV := range refSt {
				_, ok := v[refK]
				// 只替换不存在的key 存在的key以原数据为准
				if !ok {
					v[refK] = refV
				}
			}
			delete(v, refKey)
		//	字符串
		case string:
			if !strings.HasPrefix(v, prefix) || !strings.HasSuffix(v, suffix) {
				continue
			}
			refKey := v[len(prefix) : len(v)-len(suffix)]
			refData, err := loadFunc(refKey)
			if err != nil {
				return nil, err
			}
			var refSt interface{}
			err = json.Unmarshal(refData, &refSt)
			if err != nil {
				return nil, err
			}
			rootMap[k] = refSt
		}
	}
	out, err = json.Marshal(rootMap)
	return
}
