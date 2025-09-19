package logger

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

const defaultLogFileFormat = "%Y-%m-%d_%H.log"

type FileLogOption struct {
	Dir           string
	Format        string
	RotateOptions []rotatelogs.Option
}

func InitFileOutput(opts ...func(option *FileLogOption)) (writer io.Writer, err error) {
	opt := &FileLogOption{
		Dir:    "",
		Format: defaultLogFileFormat,
		RotateOptions: []rotatelogs.Option{
			rotatelogs.WithMaxAge(time.Hour * 24 * 7), // 日志保留7天
			rotatelogs.WithRotationTime(time.Hour),
		},
	}

	for _, o := range opts {
		o(opt)
	}

	if len(opt.Dir) == 0 {
		err = errors.New("invalid log directory")
		return
	}

	// 创建目录
	if err = os.MkdirAll(opt.Dir, 0775); err != nil {
		return
	}
	return rotatelogs.New(filepath.Join(opt.Dir, opt.Format), opt.RotateOptions...)
}
