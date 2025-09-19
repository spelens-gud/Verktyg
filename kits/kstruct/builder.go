package kstruct

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type (
	Constructor func(interface{}) (interface{}, error)
	CloseFunc   func(interface{}) func()
)

func MultiFieldsBuilder(cfg, st interface{}, key string, constructor Constructor, icf CloseFunc) (cf func(), err error) {
	cfgV := reflect.ValueOf(cfg)
	if cfgV.Kind() != reflect.Map {
		err = errors.New("invalid config map")
		return
	}
	v := reflect.ValueOf(st)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		err = errors.New("invalid config fields struct")
		return
	}
	var cfs []func()
	for i := 0; i < v.Type().NumField(); i++ {
		f := v.Type().Field(i)
		field := f.Tag.Get(key)
		if len(field) == 0 {
			field = f.Name
		}
		itemV := cfgV.MapIndex(reflect.ValueOf(field))
		if !itemV.IsValid() {
			err = fmt.Errorf("invalid config for %v", field)
			return nil, err
		}
		ret, err := constructor(itemV.Interface())
		if err != nil {
			return nil, err
		}
		v.Field(i).Set(reflect.ValueOf(ret))
		cfs = append(cfs, icf(ret))
	}
	cf = func() {
		for _, f := range cfs {
			f()
		}
	}
	return
}
