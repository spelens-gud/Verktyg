package kgo

import "testing"

type Car struct {
	No    string
	Type  string
	Type2 string
}

func TestGroupBy(t *testing.T) {
	testCase := []*Car{
		&Car{No: "1", Type: "cat", Type2: "cat"},
		&Car{No: "1", Type: "cat", Type2: "fly"},
		&Car{No: "2", Type: "cat", Type2: "fly"},
		&Car{No: "1", Type: "dog", Type2: "cat"},
		&Car{No: "2", Type: "dog", Type2: "fly"},
		&Car{No: "2", Type: "cat", Type2: "cat"},
		&Car{No: "1", Type: "fly", Type2: "cat"},
	}
	ret := GroupBy(testCase, func(item interface{}) interface{} {
		return item.(*Car).Type
	}, func(item interface{}) interface{} {
		return item.(*Car).Type2
	})

	t.Log(((ret["dog"]).(map[interface{}]interface{}))["cat"])

	t.Logf("%#v", ret)
	t.Logf("%#v", testCase)
}
