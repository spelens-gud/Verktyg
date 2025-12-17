package zaplog

import (
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"

	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
)

var _ ilog.Logger = &Entry{}

type (
	Entry struct {
		entry     *zap.Logger
		fieldData []interface{}
		fieldKey  []string
		tag       string
		tags      []string

		parsedFields []zap.Field
	}

	Provider struct{}
)

const (
	fieldTag  = "tag"
	fieldTags = "tags"
)

func NewEntry(writer io.Writer) ilog.Logger {
	lg := zap.New(newCore(writer), zap.AddCaller(), zap.AddCallerSkip(1))
	return &Entry{
		entry: lg,
	}
}

func (Provider) Init(writer io.Writer) ilog.Logger {
	return NewEntry(writer)
}

func sprintf(format string, a ...interface{}) string {
	if len(a) == 0 {
		return format
	}
	return fmt.Sprintf(format, a...)
}

func (e *Entry) AddCallerSkip(skip int) ilog.Logger {
	if skip == 0 {
		return e
	}
	n := e.clone()
	n.entry = n.entry.WithOptions(zap.AddCallerSkip(skip))
	return n
}

func (e *Entry) clone() *Entry {
	return &Entry{
		entry:     e.entry,
		fieldData: e.fieldData,
		fieldKey:  e.fieldKey,
		tag:       e.tag,
		tags:      e.tags,
	}
}

func (e *Entry) parseKVFields() (fields []zap.Field) {
	if len(e.parsedFields) > 0 {
		return e.parsedFields
	}

	l := len(e.fieldData)
	if l == 0 {
		return
	}
	fields = make([]zap.Field, 0, l+2)
	addedKey := make(map[string]struct{}, l+2)

	if len(e.tag) > 0 {
		addedKey[fieldTag] = struct{}{}
		fields = append(fields, zap.String(fieldTag, e.tag))
	}

	if len(e.tags) > 0 {
		addedKey[fieldTags] = struct{}{}
		fields = append(fields, zap.Strings(fieldTags, e.tags))
	}

	for j := range e.fieldKey {
		i := l - j - 1
		key := e.fieldKey[i]
		if _, added := addedKey[key]; added {
			continue
		}
		value := e.fieldData[i]
		addedKey[key] = struct{}{}
		fields = append(fields, zap.Any(key, value))
	}

	e.parsedFields = fields
	return
}

func (e *Entry) WithTag(tags ...string) ilog.Logger {
	// 去空白
	valid := 0
	for _, tag := range tags {
		if tag = strings.TrimSpace(tag); len(tag) > 0 {
			tags[valid] = tag
			valid += 1
		}
	}
	if tags = tags[:valid]; len(tags) == 0 {
		return e
	}

	var (
		entry  = e.clone()
		oldTag = entry.tag
	)

	// 主标签
	entry.tag = tags[0]

	if len(entry.tags) > 0 {
		// 已有多个Tag
		entry.tags = append(entry.tags, tags...)
	} else if len(oldTag) > 0 {
		// 已有旧Tag
		tmp := make([]string, 0, len(tags)+1)
		tmp = append(tmp, oldTag)
		tmp = append(tmp, tags...)
		entry.tags = tmp
	} else if len(tags) > 1 {
		// 单次添加多个Tag
		entry.tags = tags
	}
	return entry
}

func (e *Entry) WithField(key string, val interface{}) ilog.Logger {
	entry := e.clone()
	key = ilog.ParseLogField(key, val)
	entry.fieldData = append(entry.fieldData, val)
	entry.fieldKey = append(entry.fieldKey, key)
	return entry
}

func (e *Entry) WithFields(fields map[string]interface{}) ilog.Logger {
	if len(fields) == 0 {
		return e
	}
	entry := e.clone()
	// var omitParsed bool
	for key, val := range fields {
		//if !omitParsed {
		//	if parsedKey := ilog.ParseLogField(key, val); parsedKey == key {
		//		omitParsed = true
		//	} else {
		//		key = parsedKey
		//	}
		//}
		entry.fieldData = append(entry.fieldData, val)
		entry.fieldKey = append(entry.fieldKey, key)
	}
	return entry
}

func (e *Entry) FieldData() map[string]interface{} {
	if len(e.fieldData) == 0 {
		return nil
	}
	tmp := make(map[string]interface{}, len(e.fieldData)+2)
	if len(e.tag) > 0 {
		tmp[fieldTag] = e.tag
	}
	if len(e.tags) > 0 {
		tmp[fieldTags] = e.tags
	}
	for i := range e.fieldKey {
		tmp[e.fieldKey[i]] = e.fieldData[i]
	}
	return tmp
}

func (e *Entry) Info(msg string) {
	e.entry.Info(msg, e.parseKVFields()...)
}

func (e *Entry) Infof(msg string, args ...interface{}) {
	e.entry.Info(sprintf(msg, args...), e.parseKVFields()...)
}

func (e *Entry) Debug(msg string) {
	e.entry.Debug(msg, e.parseKVFields()...)
}

func (e *Entry) Debugf(msg string, args ...interface{}) {
	e.entry.Debug(sprintf(msg, args...), e.parseKVFields()...)
}

func (e *Entry) Warn(msg string) {
	e.entry.Warn(msg, e.parseKVFields()...)
}

func (e *Entry) Warnf(msg string, args ...interface{}) {
	e.entry.Warn(sprintf(msg, args...), e.parseKVFields()...)
}

func (e *Entry) Error(msg string) {
	e.entry.Error(msg, e.parseKVFields()...)
}

func (e *Entry) Errorf(msg string, args ...interface{}) {
	e.entry.Error(sprintf(msg, args...), e.parseKVFields()...)
}

func (e *Entry) ErrorIf(err error, messages ...string) {
	ilog.LogIfError(e.Error, err, messages...)
}

func (e *Entry) WarnIf(err error, messages ...string) {
	ilog.LogIfError(e.Warn, err, messages...)
}
