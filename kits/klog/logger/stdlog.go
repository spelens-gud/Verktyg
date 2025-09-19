package logger

import (
	"log"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger/internal"
)

func InitStandardLogger() {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(internal.StandardLogWriter())
}

func NewStandardLogger() *log.Logger {
	return log.New(internal.StandardLogWriter(), "", 0)
}
