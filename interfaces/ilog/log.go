package ilog

import (
	"context"
	"io"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type (
	LoggerProvider interface {
		Init(output io.Writer) Logger
	}

	LoggerPatch interface {
		Patch(ctx context.Context, logger Logger) Logger
	}

	Logger interface {
		// 添加Tag
		WithTag(...string) Logger
		// 添加KV
		WithField(key string, val interface{}) Logger
		// 批量添加KV
		WithFields(fields map[string]interface{}) Logger
		// 获取KV数据
		FieldData() map[string]interface{}
		// 增加调用栈深度
		AddCallerSkip(skip int) Logger

		// 各级别日志
		Info(msg string)
		Infof(msg string, args ...interface{})

		Debug(msg string)
		Debugf(msg string, args ...interface{})

		Warn(msg string)
		Warnf(msg string, args ...interface{})

		Error(msg string)
		Errorf(msg string, args ...interface{})

		// 如果err不为空 则 Warn msg + err
		WarnIf(err error, messages ...string)
		// 如果err不为空 则 Error msg + err
		ErrorIf(err error, messages ...string)
	}
)

const logTypeIdent = "_"

func LogIfError(logFunc func(string), err error, msgs ...string) {
	if err == nil {
		return
	}
	msg := strings.Join(msgs, "; ")

	if !strings.HasSuffix(msg, ": ") {
		msg += ": "
	}
	logFunc(msg + err.Error())
}

var wrapLoggerList = make(map[string]bool)

func RegisterWrapLogger(callPath string) {
	wrapLoggerList[callPath] = true
}

func ParseLogField(key string, value interface{}) (parsedKey string) {
	if len(key) == 0 {
		return
	}

	for i := 2; i < 10; i++ {
		if _, f, _, ok := runtime.Caller(i); ok {
			if wrapLoggerList[f] {
				continue
			}
			if strings.Contains(f, "sy-core@") || strings.Contains(f, "sy-core/") {
				return key
			}
		}
		break
	}

	parsedKey = key

	switch value.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int8, int16, int32, int64, uintptr,
		*uint, *uint8, *uint16, *uint32, *uint64, *int, *int8, *int16, *int32, *int64, *uintptr:
		parsedKey += logTypeIdent + "int"
	case []uint, []uint16, []uint32, []uint64, []int, []int8, []int16, []int32, []int64, []uintptr,
		*[]uint, *[]uint16, *[]uint32, *[]uint64, *[]int, *[]int8, *[]int16, *[]int32, *[]int64, *[]uintptr:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "int"
	case float32, float64,
		*float32, *float64:
		parsedKey += logTypeIdent + "float"
	case []float32, []float64,
		*[]float32, *[]float64:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "float"
	case []byte,
		*[]byte:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "byte"
	case string,
		*string:
		parsedKey += logTypeIdent + "string"
	case []string,
		*[]string:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "string"
	case complex64, complex128,
		*complex64, *complex128:
		parsedKey += logTypeIdent + "complex"
	case []complex64, []complex128,
		*[]complex64, *[]complex128:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "complex"
	case time.Time,
		*time.Time:
		parsedKey += logTypeIdent + "time"
	case []time.Time,
		*[]time.Time:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "time"
	case map[string]interface{}, map[int]interface{},
		*map[string]interface{}, *map[int]interface{},
		*interface{}:
		parsedKey += logTypeIdent + "object"
	case []interface{}:
		parsedKey += logTypeIdent + "array" + logTypeIdent + "object"
	case nil:
		parsedKey += logTypeIdent + "nil"
	default:
		vt := reflect.TypeOf(value)
		parsedKey += parseLogFieldReflect(vt)
	}
	return
}

func parseLogFieldReflect(vt reflect.Type) (parsedKeySuffix string) {
	switch vt.Kind() {
	case reflect.Ptr:
		return parseLogFieldReflect(vt.Elem())
	case reflect.Slice:
		return logTypeIdent + "array" + parseLogFieldReflect(vt.Elem())
	case reflect.Struct, reflect.Map, reflect.Interface:
		return logTypeIdent + "object"
	default:
		k := vt.Kind().String()
		switch {
		case k == "interface", k == "struct", k == "map":
			return "object"
		case strings.Contains(k, "int"):
			return logTypeIdent + "int"
		case strings.Contains(k, "float"):
			return logTypeIdent + "float"
		case strings.Contains(k, "complex"):
			return logTypeIdent + "complex"
		default:
			return logTypeIdent + k
		}
	}
}
