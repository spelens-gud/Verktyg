package go2ts

import (
	"github.com/spelens-gud/Verktyg.git/interfaces/idoc"
)

type Parser struct {
	Types []*idoc.TsType
}

func ParseJsonSchema(in interface{}) (re *idoc.JsSchema) {
	return TsTypes2JsonSchema(ParseTsTypes(in))
}

func TsTypes2JsonSchema(p []*idoc.TsType) (re *idoc.JsSchema) {
	if len(p) == 0 {
		return
	}
	re = &idoc.JsSchema{Properties: idoc.Definitions{}, Definitions: idoc.Definitions{}}
	tm := map[string]*idoc.JsSchema{}
	for i, t := range p {
		var nt *idoc.JsSchema
		if i == len(p)-1 {
			nt = re
		} else {
			nt = &idoc.JsSchema{Properties: idoc.Definitions{}}
		}
		tm[t.Name] = nt
		for _, f := range t.Fields {
			p := new(idoc.JsSchema)
			nt.Properties[f.Name] = p
			if f.IsArray {
				p.Type = "array"
				p.Items = new(idoc.JsSchema)
				p = p.Items
			}
			if tm[f.Type] != nil {
				*p = *tm[f.Type]
				continue
			}
			p.Type = f.Type
			p.Description = f.Comment
			p.Title = f.Name
		}
	}
	return
}
