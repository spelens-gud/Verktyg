package iconfig

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"git.bestfulfill.tech/devops/go-core/kits/kgo"
)

var (
	execPath = os.Args[0]

	execAbsoluteDir = func() string {
		ret, err := filepath.Abs(filepath.Dir(execPath))
		if err != nil {
			panic(err)
		}
		return ret
	}()

	appName   = &configString{Init: getApplicationAppName}
	namespace = &configString{Init: getApplicationNameSpace}
	hostname  = &configString{Init: getHostname}
	execMd5   = &configString{Init: getExecMd5}
)

// 获取执行文件目录
func ExecDir() string { return execAbsoluteDir }

// 获取执行文件名
func ExecName() string { return filepath.Base(execPath) }

// 获取Hostname
func HostName() string { return hostname.Get() }

// 获取应用命名空间
func GetApplicationNameSpace() string { return namespace.Get() }

// 设置应用命名空间
func SetApplicationNameSpace(str string) { namespace.Set(str) }

// 获取应用名
func GetApplicationAppName() string { return appName.Get() }

// 设置应用名
func SetApplicationAppName(str string) { appName.Set(str) }

// 获取应用完整名 格式: $namespace:$app
func GetApplicationName() string { return GetApplicationNameSpace() + ":" + GetApplicationAppName() }

// 获取应用服务名 格式: $app.$namespace
func GetApplicationServiceName() string {
	return GetApplicationAppName() + "." + GetApplicationNameSpace()
}

func initFormat(in string) string {
	return strings.ReplaceAll(strings.ToLower(in), "_", "-")
}

func getHostname() (n string) { n, _ = os.Hostname(); return n }

func getApplicationNameSpace() (n string) {
	if n = os.Getenv(EnvKeyContainerNamespace); len(n) == 0 {
		n = "unknown"
	}
	return initFormat(n)
}

func getApplicationAppName() (n string) {
	defaultFileName := "unknown-app-"
	if hash := execMd5.Get(); len(hash) >= 8 {
		defaultFileName += hash[:8]
	} else {
		defaultFileName += uuid.New().String()[:8]
	}
	return initFormat(kgo.InitStrings(os.Getenv(EnvKeyContainerAppName), defaultFileName))
}

func getExecMd5() (n string) {
	md5Hash, _ := getFileMd5(execPath)
	return md5Hash
}

func getFileMd5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
