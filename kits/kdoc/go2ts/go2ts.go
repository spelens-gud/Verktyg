package go2ts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/spelens-gud/Verktyg/interfaces/idoc"
	"github.com/spelens-gud/Verktyg/kits/kdoc"
)

const DefaultTsTemplate = `
{{ range .Types }}export interface {{.Name}} {
	{{ range .Fields }}{{ .Name }} {{ if eq .Option true }}?{{ end }}: {{ .TsType }}{{ if eq .IsArray true }}[]{{ end }} ; // {{.Comment}}
	{{ end }}
}

{{ end }}
`

var DefaultTemplate = template.Must(template.New("template").Parse(DefaultTsTemplate))

func ParseStruct2TsInterface(in interface{}) (out string, err error) {
	var bf bytes.Buffer
	err = DefaultTemplate.Execute(&bf, &Parser{Types: ParseTsTypes(in)})
	out = bf.String()
	return
}

func ParseTsTypes(in interface{}) (tps []*idoc.TsType) {
	p := &Parser{}
	p.parseInterface(in, &idoc.TsField{})
	tps = p.Types
	return
}

func (p *Parser) parseInterface(in interface{}, field *idoc.TsField) {
	switch in.(type) {
	case time.Time, *time.Time:
		field.Type = "Date"
	case bool, *bool:
		field.Type = "boolean"
	case string, []byte, *string:
		field.Type = "string"
	case int, int64, int8, int16, int32, uint, uint8, uint16, uint32, uint64, float64, float32, json.Number,
		*int, *int64, *int8, *int16, *int32, *uint, *uint8, *uint16, *uint32, *uint64, *float64, *float32, *json.Number:
		field.Type = "number"
	default:
		s := reflect.ValueOf(in)
		for s.Kind() == reflect.Ptr || s.Kind() == reflect.Interface {
			s = s.Elem()
		}
		switch s.Kind() {
		case reflect.Map:
			p.parseMap(s.Interface(), field)
		case reflect.Slice:
			p.parseArray(s.Interface(), field)
		case reflect.Struct:
			m := kdoc.GetStructFieldsDoc(s.Interface())
			p.parseStrut(s.Interface(), m, field)
		default:
			field.Type = "any"
		}
	}
}

func (p *Parser) parseMap(in interface{}, field *idoc.TsField) {
	s := reflect.ValueOf(in)
	isAllSame := true
	var t reflect.Type
	for _, k := range s.MapKeys() {
		if len(s.MapKeys()) == 1 {
			break
		}
		if t == nil {
			if s.MapIndex(k).Kind() == reflect.Interface {
				if s.MapIndex(k).IsNil() {
					continue
				}
				t = s.MapIndex(k).Elem().Type()
			} else {
				t = s.MapIndex(k).Type()
			}
		} else {
			var t2 reflect.Type
			if s.MapIndex(k).Kind() == reflect.Interface {
				if s.MapIndex(k).IsNil() {
					continue
				}
				t2 = s.MapIndex(k).Elem().Type()
			} else {
				t2 = s.MapIndex(k).Type()
			}
			if t != t2 {
				isAllSame = false
				break
			}
		}
	}
	if isAllSame {
		v := reflect.New(s.Type().Elem()).Elem().Interface()
		p.parseInterface(v, field)
		var keyField idoc.TsField
		p.parseInterface(reflect.New(s.Type().Key()).Elem().Interface(), &keyField)
		field.Type = fmt.Sprintf("{ [ key : %s ] : %s }", keyField.Type, field.Type)
	} else {
		// todo:map不同key
		if s.Type().Key().Kind() != reflect.String {
			return
		}
		mapTpy := new(idoc.TsType)
		mapTpy.Name = field.Name
		for _, k := range s.MapKeys() {
			str, _ := k.Interface().(string)
			newMapField := &idoc.TsField{
				Name: str,
			}
			p.parseInterface(s.MapIndex(k).Interface(), newMapField)
			mapTpy.Fields = append(mapTpy.Fields, newMapField)
		}
	}
}

func (p *Parser) parseStrut(in interface{}, commentDesc map[string]string, field *idoc.TsField) {
	tpy := new(idoc.TsType)
	name := reflect.TypeOf(in).Name()
	tpy.Name = name
	field.Type = name
	p.appendStrutField(in, tpy, commentDesc)
	p.Types = append(p.Types, tpy)
}

func (p *Parser) appendStrutField(in interface{}, typ *idoc.TsType, commentDesc map[string]string) {
	s := reflect.ValueOf(in)
	for s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < s.Type().NumField(); i++ {
		f := s.Type().Field(i)
		comment := commentDesc[f.Name]
		if len(f.PkgPath) > 0 {
			continue
		}
		if f.Anonymous {
			// nolint
			if !(f.Type.Kind() == reflect.Ptr && s.Field(i).IsNil()) {
				p.appendStrutField(s.Field(i).Interface(), typ, commentDesc)
			}
			continue
		}
		var fieldName string
		jk := f.Tag.Get("json")
		newTsField := new(idoc.TsField)
		if len(jk) == 0 {
			fieldName = f.Name
		} else if jk == "-" {
			continue
		} else {
			if strings.Contains(jk, ",omitempty") {
				if s.Field(i).IsZero() {
					continue
				}
				newTsField.Option = true
			}
			fieldName = strings.Split(jk, ",")[0]
		}
		newTsField.Name = fieldName
		newTsField.Comment = comment
		p.parseInterface(s.Field(i).Interface(), newTsField)
		newTsField.Comment = strings.TrimSpace(newTsField.Comment)
		if newTsField.Type != "" {
			typ.Fields = append(typ.Fields, newTsField)
		}
	}
}

func (p *Parser) parseArray(in interface{}, field *idoc.TsField) {
	field.IsArray = true
	switch t := in.(type) {
	case []interface{}:
		if len(t) > 0 {
			p.parseInterface(t[0], field)
		}
	default:
		value := reflect.ValueOf(in)
		if value.Len() == 1 {
			p.parseInterface(value.Index(0).Interface(), field)
			return
		}
		rTyp := value.Type().Elem()
		for rTyp.Kind() == reflect.Ptr {
			rTyp = rTyp.Elem()
		}
		p.parseInterface(reflect.New(rTyp).Elem().Interface(), field)
	}
}
