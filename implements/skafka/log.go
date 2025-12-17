package skafka

import (
	"fmt"
	"runtime"

	"github.com/Shopify/sarama"

	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

func init() {
	newLogger := logger.NewStandardLogger()
	newLogger.SetPrefix("[SARAMA] ")
	sarama.Logger = newLogger
	sarama.PanicHandler = func(r interface{}) {
		buf := make([]byte, 64<<10)
		buf = buf[:runtime.Stack(buf, false)]
		info := fmt.Sprintf("[PANIC] recovered: %s\n%s", r, buf)
		fmt.Print(info)
		logger.FromBackground().WithTag("KAFKA_PANIC").Error(info)
	}
}
