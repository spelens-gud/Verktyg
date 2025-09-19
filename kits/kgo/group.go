package kgo

import (
	"reflect"
)

type GroupResult map[interface{}]interface{}

func (r GroupResult) GetItems(keys ...interface{}) ([]interface{}, bool) {
	ptr := r
	for _, k := range keys {
		switch v := ptr[k].(type) {
		case []interface{}:
			return v, true
		case GroupResult:
			ptr = v
		default:
			return nil, false
		}
	}
	return nil, false
}

func GroupBy(list interface{}, by ...func(item interface{}) interface{}) (ret GroupResult) {
	if len(by) == 0 {
		return nil
	}

	parse := func(list interface{}, b func(item interface{}) interface{}) GroupResult {
		v := make(GroupResult)
		fv := reflect.ValueOf(list)
		if fv.Kind() != reflect.Slice {
			panic("invalid list type")
		}
		for i := 0; i < fv.Len(); i++ {
			item := fv.Index(i).Interface()
			key := b(item)
			l, _ := v[key].([]interface{})
			if l == nil {
				l = make([]interface{}, 0)
			}
			v[key] = append(l, item)
		}
		return v
	}

	if ret = parse(list, by[0]); len(by) > 1 {
		for k, v := range ret {
			ret[k] = GroupBy(v.([]interface{}), by[1:]...)
		}
	}
	return
}
