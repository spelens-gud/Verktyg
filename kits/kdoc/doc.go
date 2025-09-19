package kdoc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"reflect"
	"strings"
)

func GetStructFieldsDoc(in interface{}) (res map[string]string) {
	return getStructFieldsDoc(reflect.TypeOf(in))
}

func getStructFieldsDoc(tpy reflect.Type) (res map[string]string) {
	for tpy.Kind() == reflect.Ptr {
		tpy = tpy.Elem()
	}
	if tpy.Kind() != reflect.Struct {
		return
	}
	importPath, ok := getImportCodePath(tpy.PkgPath())
	if !ok {
		return
	}
	d, err := ioutil.ReadDir(importPath)
	if err != nil {
		return
	}
	res = make(map[string]string)
	for _, d := range d {
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			continue
		}
		fileData, _ := ioutil.ReadFile(path.Join(importPath, d.Name()))
		f, err := parser.ParseFile(token.NewFileSet(), "", fileData, parser.ParseComments)
		if err != nil {
			continue
		}
		get := f.Scope.Lookup(tpy.Name())
		if get == nil || get.Kind != ast.Typ {
			continue
		}
		sp, ok := get.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := sp.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			continue
		}
		for i, f := range st.Fields.List {
			// 匿名函数
			if f.Names == nil {
				for k, v := range getStructFieldsDoc(tpy.Field(i).Type) {
					res[k] = v
				}
				continue
			}
			fieldName := f.Names[0].Name
			if f.Comment != nil {
				res[fieldName] += strings.Replace(f.Comment.Text(), "\n", " ", -1) + " "
			}
			if f.Doc != nil {
				res[fieldName] += strings.Replace(f.Doc.Text(), "\n", " ", -1) + " "
			}
			res[fieldName] = strings.TrimSpace(res[fieldName])
			if len(res[fieldName]) == 0 {
				delete(res, fieldName)
			}
		}
	}
	return
}
