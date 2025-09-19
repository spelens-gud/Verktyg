package kurl

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

var defaultTag = []string{"url"}

// omitEmptyTag为true时  没有该tag则会被省略 否则以字段原名为key
func parse2UrlValues(in interface{}, omitEmptyTag, omitIfMap bool, tags ...string) (v url.Values) {
	if in == nil {
		return
	}
	t := reflect.TypeOf(in)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Map:
		if omitIfMap {
			return nil
		}
		return ParseMap2UrlValues(in)
	case reflect.Struct:
		return ParseStruct2UrlValues(in, omitEmptyTag, tags...)
	}
	return nil
}

func ParseMap2UrlValues(in interface{}) (v url.Values) {
	if in == nil {
		return nil
	}
	switch m := in.(type) {
	case map[string]string:
		v = make(url.Values)
		for k, value := range m {
			v.Add(k, value)
		}
	case map[string]interface{}:
		v = make(url.Values)
		for k, value := range m {
			str, e := cast.ToStringE(value)
			if e != nil {
				str = fmt.Sprint(value)
			}
			v.Add(k, str)
		}
	default:
		value := reflect.Indirect(reflect.ValueOf(in))
		t := value.Type()
		if t.Kind() != reflect.Map {
			panic("map required")
		}
		if valueKind := t.Elem().Kind(); valueKind > reflect.Float64 || valueKind == reflect.Invalid {
			if valueKind != reflect.String {
				return nil
			}
		}
		v = make(url.Values)
		rV := reflect.Indirect(value)
		for _, k := range rV.MapKeys() {
			v.Add(fmt.Sprint(k.Interface()), fmt.Sprint(rV.MapIndex(k).Interface()))
		}
	}
	return
}

func Parse2UrlValues(in interface{}, omitIfMap bool, tags ...string) (v url.Values) {
	return parse2UrlValues(in, false, omitIfMap, tags...)
}

func Parse2UrlValuesOmitEmptyTag(in interface{}, omitIfMap bool, tags ...string) (v url.Values) {
	return parse2UrlValues(in, true, omitIfMap, tags...)
}

func ParseStruct2UrlValues(in interface{}, omitEmptyTag bool, tags ...string) (v url.Values) {
	if in == nil {
		return
	}
	if len(tags) == 0 {
		tags = defaultTag
	}
	rV := reflect.ValueOf(in)
	v = make(url.Values)
	parseStruct2UrlValues(rV, omitEmptyTag, v, tags...)
	return
}

func parseStruct2UrlValues(rV reflect.Value, omitEmptyTag bool, v url.Values, tags ...string) {
	if rV = reflect.Indirect(rV); rV.Kind() != reflect.Struct {
		panic("struct required")
	}

	for i := 0; i < rV.NumField(); i++ {
		var (
			typ    = rV.Type().Field(i)
			rawKey string
		)

		// 跳过不暴露的字段
		if typ.PkgPath != "" && !typ.Anonymous {
			continue
		}

		// 找到可用的tag
		for _, tag := range tags {
			var ok bool
			if rawKey, ok = typ.Tag.Lookup(tag); ok {
				break
			}
		}

		// 如果找出来tag为- 则跳过
		if rawKey == "-" {
			continue
		}

		tagKey, options := parseTag(rawKey)
		// 如果没有key
		if len(tagKey) == 0 {
			// 若省略空tag 则跳过
			if omitEmptyTag && !typ.Anonymous {
				continue
			}
			// 否则以字段名为key
			tagKey = typ.Name
		}

		// 字段值
		value := reflect.Indirect(rV.Field(i))
		// 如果是匿名字段且为结构体 则深入一层
		if typ.Anonymous && value.Kind() == reflect.Struct {
			parseStruct2UrlValues(value, omitEmptyTag, v, tags...)
			continue
		}

		// 如果omitempty且值为空
		if options.Contains("omitempty") && value.IsZero() {
			continue
		}

		// 添加到url里
		fv := value.Interface()
		if str, err := cast.ToStringE(fv); err == nil {
			v.Add(tagKey, str)
		} else {
			v.Add(tagKey, fmt.Sprint(fv))
		}
	}
}

func parseTag(tag string) (string, tagOptions) {
	s := strings.Split(tag, ",")
	return s[0], s[1:]
}

type tagOptions []string

func (o tagOptions) Contains(option string) bool {
	for _, s := range o {
		if s == option {
			return true
		}
	}
	return false
}
