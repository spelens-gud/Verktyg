package kdb

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var envTimeout = os.Getenv("PING_DB_TIMEOUT_SECONDS")

func PingDB(ping func() error, name string, addr string) (err error) {
	logger.FromBackground().Infof("正在检查连接 %s [ %s ]", name, addr)

	errChan := make(chan error)
	go func() {
		errChan <- ping()
	}()
	timeout := 3 * time.Second
	if timeoutInt, _ := strconv.Atoi(envTimeout); timeoutInt > 0 {
		timeout = time.Second * time.Duration(timeoutInt)
	}

	select {
	case err = <-errChan:
	case <-time.After(timeout):
		err = fmt.Errorf("检查连接超时 [ %s ]", timeout)
	}

	if err != nil {
		logger.FromBackground().Errorf("检查连接 %s 失败 [ %s ]: %v", name, addr, err)
		err = errors.WithStack(err)
		return err
	}

	logger.FromBackground().Infof("检查连接 %s 成功 [ %s ]", name, addr)
	return
}
