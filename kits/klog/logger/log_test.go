package logger_test

import (
	"context"
	"io"
	"log"
	"os"
	"testing"

	"github.com/spelens-gud/Verktyg.git/implements/zaplog"
	"github.com/spelens-gud/Verktyg.git/interfaces/ilog"
	"github.com/spelens-gud/Verktyg.git/interfaces/itrace"
	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

var ctx = context.Background()

func init() {
}

func BenchmarkInitLogger(b *testing.B) {
	b.ReportAllocs()
	logger.SetLevel(ilog.Error)
	logger.SetOutput(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.FromContext(ctx).Info("1")
	}
}

func BenchmarkStandardLogger(b *testing.B) {
	b.ReportAllocs()
	logger.InitStandardLogger()
	logger.SetLevel(ilog.Info)
	logger.SetOutput(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Print("xx")
	}
}

func BenchmarkLoggerDoLog(b *testing.B) {
	b.ReportAllocs()
	logger.SetLevel(ilog.Info)
	logger.SetOutput(io.Discard)
	b.ResetTimer()
	lg := logger.FromContext(ctx)
	for i := 0; i < b.N; i++ {
		lg.Info("1")
	}
}

func BenchmarkLogZap(b *testing.B) {
	b.ReportAllocs()
	e := zaplog.NewEntry(io.Discard)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Info("1")
	}
}

func TestLogWriter(t *testing.T) {
	output := logger.Writer()
	_, _ = output.Write([]byte("1\n"))
	{
		if err := logger.SetFileOutput(func(option *logger.FileLogOption) {
			option.Dir = "./"
		}); err != nil {
			t.Fatal(err)
		}
		_, _ = output.Write([]byte("2\n"))
		logger.SetOutput(os.Stdout)
		_, _ = output.Write([]byte("3\n"))
	}
}

func TestLog(t *testing.T) {
	logger.NewStandardLogger().Print("7")
	logger.FromContext(ctx).Infof("1")

	logger.SetLevel(ilog.Error)
	logger.FromContext(ctx).Debugf("2")
	logger.FromContext(ctx).Errorf("3")

	logger.SetLevel(ilog.Info)
	logger.FromContext(ctx).Debugf("4")
	logger.FromContext(ctx).Infof("5")

	{
		if err := logger.SetFileOutput(func(option *logger.FileLogOption) {
			option.Dir = "./"
		}); err != nil {
			t.Fatal(err)
		}
		logger.FromContext(itrace.WithContext(ctx, "test-id")).Infof("6")
	}

	logger.SetLevel(0)
	logger.NewStandardLogger().Print("7")
	logger.FromContext(ctx).Infof("8")
}
