package logger

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/spelens-gud/Verktyg.git/interfaces/iconfig"
)

const defaultLogFileFormat = "%Y-%m-%d_%H.log"

type FileLogOption struct {
	Dir           string
	Format        string
	RotateOptions []rotatelogs.Option
}

// 设置日志输出文件目录
func SetFileOutput(opts ...func(option *FileLogOption)) (err error) {
	output, err := initFileOutput(opts...)
	if err != nil {
		return
	}
	// 非生产环境 日志同时输出控制台
	if !iconfig.GetEnv().IsProduction() {
		output = io.MultiWriter(os.Stdout, output)
	}
	SetOutput(output)
	return
}

func initFileOutput(opts ...func(option *FileLogOption)) (writer io.Writer, err error) {
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
