package dotenv

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

var (
	ImportMe struct{}

	logger = log.New(os.Stderr, "[SYCORE] ", log.LstdFlags)
)

const envFileName = ".env"

func init() {
	// 初始化地址
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Fatalf("load env filepath failed: %v", err)
	}
	before := len(os.Environ())
	// 读 env 文件
	file := filepath.Join(dir, envFileName)
	// 第一次从os.Args[0]目录读
	if err = godotenv.Overload(file); err != nil {
		// 失败之后重试在当前目录读
		logger.Printf("load [ %s ] failed: %v,try [ %s ]", file, err, envFileName)
		if err = godotenv.Overload(envFileName); err != nil {
			if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
				logger.Printf("load [ %s ] failed: %v", envFileName, err)
				return
			}
			logger.Fatalf("load [ %s ] failed: %v", envFileName, err)
		}
	}
	after := len(os.Environ())
	logger.Printf("load [ %s ] success,[ %d ] envs loaded", envFileName, after-before)
}
