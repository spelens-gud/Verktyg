package zaplog_test

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"testing"

	"go.uber.org/zap"

	"github.com/spelens-gud/Verktyg.git/implements/zaplog"
)

var output io.Writer

func init() {
	output = io.Discard
}

func TestEntry_Info(t *testing.T) {
	zaplog.NewEntry(output).WithField("ttt", "1111").Info("111")
}

func BenchmarkLog(t *testing.B) {
	l := zaplog.NewEntry(output)
	t.ReportAllocs()
	t.ResetTimer()
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.WithFields(map[string]interface{}{
				"xbczx":   "gads",
				"1212":    "gads",
				"2353223": "gads",
			})
		}
	})
}

func TestField(t *testing.T) {
	l := zaplog.NewEntry(os.Stdout)
	l1 := l.WithField("1", "1111").WithFields(map[string]interface{}{
		"xbc":   "42",
		"xbc2":  "42",
		"xbc42": "42",
	})
	l2 := l1.WithField("2", "1111")
	l3 := l2.WithField("3", "1111")
	l4 := l3.WithField("4", "1111")
	l5 := l4.WithField("5", "1111")
	l1.Info("xx")
	l1.Info("xx22")
	l2.Info("xx2")
	l3.Info("xx3")
	l4.Info("xx4")
	l5.Info("xx5")

}

func TestEntry_Tag(t *testing.T) {
	l := zaplog.NewEntry(output)
	l2 := l.WithTag("111")
	fmt.Println(l)
	fmt.Println(l2)
	l.Info("222")
	l2.Info("333")
	l2.WithTag("4444").Info("222")
	l2.Info("555")
}

func TestWriteLog(t *testing.T) {
	runtime.GOMAXPROCS(1)
	rawL, _ := zap.NewProduction()
	rawL.Info("")
	wg := new(sync.WaitGroup)
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = os.Stdout.Write([]byte("xx"))
		}()
	}
	wg.Wait()
}
