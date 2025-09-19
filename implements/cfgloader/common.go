package cfgloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"git.bestfulfill.tech/devops/go-core/interfaces/iconfig"
	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
	"git.bestfulfill.tech/devops/go-core/internal/incontext"
	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

var (
	pkgPath = func() string {
		_, f, _, _ := runtime.Caller(0)
		return filepath.Dir(f)
	}()

	configCachePath = filepath.Join(iconfig.ExecDir(), "acm-cache", iconfig.ExecName()+".config.json")
)

const (
	defaultConfigTemplateName = "config.template.json"
	defaultAcmConfigDir       = "/acm-config-data"
	defaultLocalConfigName    = "config.local.json"
	defaultOldLocalConfigName = "acm.local.json"
	keyInWrapped              = incontext.Key("config_loader.in_wrapped")
)

func NewConfigLoaderFromEnv(extLoaders ...iconfig.ConfigLoader) iconfig.ConfigLoader {
	loaders := make(multi, 0)

	if iconfig.GetEnv().IsDevelopment() {
		loaders = append(loaders, NewFileLoader(defaultLocalConfigName, defaultOldLocalConfigName))
	}

	if newNacosOptionFromEnv().Validate() == nil {
		loaders = append(loaders, NewAcmLoader())
	}

	if _, e := os.Stat(defaultAcmConfigDir); e == nil {
		loaders = append(loaders, NewFileLoader(filepath.Join(defaultAcmConfigDir, iconfig.GetApplicationAppName())))
	}

	return NewMultiLoader(append(loaders, extLoaders...)...)
}

func getCaller() (dir string, ok bool) {
	for i := 1; i < 10; i++ {
		_, callFile, _, ok := runtime.Caller(i)
		if !ok {
			break
		}
		if strings.HasPrefix(callFile, pkgPath) || strings.Contains(callFile, "/sync/") {
			continue
		}
		return filepath.Dir(callFile), true
	}
	return
}

func getLogger(ctx context.Context, loader iconfig.ConfigLoader) ilog.Logger {
	return logger.FromContext(ctx).WithField("loader", fmt.Sprintf("%T", loader))
}

func wrapLoad(ctx context.Context, configRoot interface{}, load func(ctx context.Context) (err error)) (err error) {
	if keyInWrapped.Value(ctx) != nil {
		return load(ctx)
	}
	ctx = keyInWrapped.WithValue(ctx, struct{}{})

	lg := logger.FromContext(ctx)

	if iconfig.GetEnv().IsDevelopment() {
		if dir, ok := getCaller(); ok {
			// 写入空结构体模板文件
			f := filepath.Join(dir, defaultConfigTemplateName)
			if e := writeConfig(configRoot, f); e != nil {
				lg.Errorf("write config template error: %v", e)
				return
			}
			lg.WithField("file", f).Infof("update config template")
		}
	}

	if iconfig.GetEnv().IsProduction() {
		defer func() {
			if err == nil {
				// 写缓存文件
				if e := writeConfig(configRoot, configCachePath); e != nil {
					logger.FromBackground().WithField("path", configCachePath).
						Warnf("write config cache error: %v", e)
				}
				return
			}

			// 降级缓存文件
			lg.Warnf("load config error: %v,try fallback cache", err)
			if err = NewFileLoader(configCachePath).LoadContext(ctx, configRoot); err != nil {
				lg.WithField("cache", configCachePath).Errorf("fallback cache error: %v", err)
			}
		}()
	}

	if err = load(ctx); err != nil {
		return
	}
	return
}

func writeConfig(configRoot interface{}, file string) (err error) {
	data, err := json.MarshalIndent(configRoot, "", "\t")
	if err != nil {
		return
	}
	if err = os.MkdirAll(filepath.Dir(file), 0775); err != nil {
		return
	}
	return os.WriteFile(file, data, 0664)
}

func loadFromBytes(ctx context.Context, f func() ([]byte, error), configRoot interface{}, typ iconfig.ConfigType) (err error) {
	lg := logger.FromContext(ctx)
	configData, err := f()
	if err != nil {
		lg.Errorf("load config error: %v", err)
		return
	}

	if err = typ.Unmarshal(configData, configRoot); err != nil {
		lg.Errorf("unmarshal config error: %v", err)
		return
	}
	lg.Infof("config loaded")
	return
}
