package enumx

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	skipDir = []string{
		"vendor",
		".idea",
		".git",
		"cmd",
		"api",
		".DS_Store",
		".githooks",
	}
	regexpEnum = regexp.MustCompile(`@enum\((.+?)\)`)
)

type (
	Enum struct {
		Name string `json:"name"`
		Desc string `json:"desc"`
		List []Item `json:"list"`
	}

	Item struct {
		Key  string `json:"key"`
		Desc string `json:"desc"`
	}

	Enumap map[string]Enum

	Enumer interface {
		Get(key string) (value Enum, ok bool)

		GetAll() (value Enumap)
	}

	enumer struct {
		m Enumap
		*sync.Mutex
	}
)

func LoadFrom(b []byte) (e Enumer, err error) {

	var m Enumap
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	return &enumer{
		m:     m,
		Mutex: new(sync.Mutex),
	}, nil
}

func (e *enumer) Get(key string) (value Enum, ok bool) {
	e.Lock()
	value, ok = e.m[key]
	e.Unlock()
	return
}

func (e enumer) GetAll() (value Enumap) {
	return e.m
}

func (e Enum) GetKey(key string) (value string, ok bool) {
	for i := range e.List {
		if e.List[i].Key == key {
			return e.List[i].Desc, true
		}
	}
	return
}

func (e Enum) String() string {
	var d []string
	for _, l := range e.List {
		d = append(d, l.Key+"-"+l.Desc)
	}
	return e.Desc + ":(" + strings.Join(d, ",") + ")"
}

func (em Enumap) Replace(in string) string {
	return em.ReplaceWith(in, func(e Enum) string {
		return e.String()
	})
}

func (em Enumap) ReplaceWith(in string, repFunc func(e Enum) string) string {
	founds := regexpEnum.FindAllStringSubmatch(in, -1)
	if len(founds) == 0 {
		return in
	}
	for _, fs := range founds {
		if len(fs) != 2 {
			continue
		}
		e, ok := em[fs[1]]
		if !ok {
			continue
		}
		in = strings.ReplaceAll(in, fs[0], repFunc(e))
	}
	return in
}

func isSkip(dir string) bool {
	for _, d := range skipDir {
		if dir == d {
			return true
		}
	}
	return false
}

func Search(file string) (enumap Enumap) {
	enumap = make(map[string]Enum)
	_ = filepath.Walk(file, func(path string, info os.FileInfo, err error) error {
		_ = err
		if (isSkip(info.Name())) && info.IsDir() {
			return filepath.SkipDir
		}
		if len(info.Name()) > 3 && info.Name()[len(info.Name())-3:] != ".go" || info.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		fs := token.NewFileSet()
		f, _ := parser.ParseFile(fs, path, b, parser.ParseComments)
		for _, d := range f.Decls {
			decl, ok := d.(*ast.GenDecl)
			if !ok || decl.Tok != token.CONST || decl.Doc == nil || !regexpEnum.MatchString(decl.Doc.Text()) || len(decl.Specs) == 0 {
				continue
			}
			text := decl.Doc.Text()
			matches := regexpEnum.FindStringSubmatch(text)
			if len(matches) != 2 {
				continue
			}
			desc := strings.ReplaceAll(text, matches[0], "")
			e := Enum{
				List: make([]Item, 0, len(decl.Specs)),
				Name: matches[1],
				Desc: strings.TrimSpace(strings.ReplaceAll(desc, "\n", " ")),
			}
			for _, sp := range decl.Specs {
				v, ok := sp.(*ast.ValueSpec)
				if !ok || len(v.Names) != 1 || len(v.Values) != 1 {
					continue
				}
				ds := strings.Split(v.Doc.Text(), "\n")
				cs := strings.Split(v.Comment.Text(), "\n")
				desc := strings.TrimSpace(strings.Join(append(ds, cs...), " "))
				str := v.Values[0]
				e.List = append(e.List, Item{
					Key:  string(b[str.Pos()-1 : str.End()-1]),
					Desc: desc,
				})
			}
			enumap[e.Name] = e
		}
		return nil
	})
	return
}
