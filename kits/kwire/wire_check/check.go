package wire_check

import (
	"fmt"
	"reflect"
)

func Check(in interface{}) {
	check(reflect.ValueOf(in))
}

func check(in reflect.Value) {
	for in.Kind() == reflect.Interface || in.Kind() == reflect.Ptr {
		if in.IsNil() {
			panic(fmt.Errorf("empty interface:%v:%v", in.Type().PkgPath(), in.Type()))
		}
		in = in.Elem()
	}
	if in.Kind() != reflect.Struct {
		return
	}
	nf := in.NumField()
	t := in.Type()
	for i := 0; i < nf; i++ {
		if in.Field(i).Kind() == reflect.Interface {
			func() {
				defer func() {
					e := recover()
					if e != nil {
						panic(fmt.Errorf("empty fields %v:%v:%v\n%v", t.Field(i).Name, t, t.PkgPath(), e))
					}
				}()
				check(in.Field(i))
			}()
		}
	}
}
