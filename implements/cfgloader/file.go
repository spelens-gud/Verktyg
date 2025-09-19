package cfgloader

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var _ iconfig.ConfigLoader = &files{}

type files []string

// 通过静态文件加载配置  支持多个路径
func NewFileLoader(filePaths ...string) iconfig.ConfigLoader {
	return files(filePaths)
}

func (fs files) Load(configRoot interface{}) (err error) {
	return fs.LoadContext(context.Background(), configRoot)
}

func (fs files) MustLoad(configRoot interface{}) { must(fs.Load(configRoot)) }

func (fs files) LoadContext(ctx context.Context, configRoot interface{}) (err error) {
	return wrapLoad(ctx, configRoot, func(ctx context.Context) (err error) {
		var extDirPath []string
		for _, f := range fs {
			if f = strings.TrimSpace(f); filepath.IsAbs(f) {
				continue
			}
			// 添加调用来源当前目录
			if callerDir, ok := getCaller(); ok {
				nf := filepath.Join(callerDir, f)
				if _, e := os.Stat(nf); e == nil {
					extDirPath = append(extDirPath, nf)
				}
			}

			// 添加二进制当前目录
			nf := filepath.Join(iconfig.ExecDir(), f)
			if _, e := os.Stat(nf); e == nil {
				extDirPath = append(extDirPath, nf)
			}
		}

		for _, f := range append(fs, extDirPath...) {
			if loadFromFile(logger.WithContext(ctx, getLogger(ctx, fs)), f, configRoot) == nil {
				return
			}
		}
		return errors.New("load config from all files failed")
	})
}

func loadFromFile(ctx context.Context, f string, configRoot interface{}) (err error) {
	ctx = logger.WithField(ctx, "file", f)

	var typ iconfig.ConfigType

	switch {
	case strings.HasSuffix(f, ".json"):
		typ = iconfig.ConfigTypeJson
	case strings.HasSuffix(f, ".yml"), strings.HasSuffix(f, ".yaml"):
		typ = iconfig.ConfigTypeYaml
	}

	return loadFromBytes(ctx, func() ([]byte, error) {
		return ioutil.ReadFile(f)
	}, configRoot, typ)
}
