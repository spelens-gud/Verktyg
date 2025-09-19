package gormctx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"git.bestfulfill.tech/devops/go-core/kits/klog/logger"
)

type gormLogger struct {
	db         *ctxConn
	detailMode bool
}

func (g *gormLogger) Print(v ...interface{}) {
	if len(v) == 0 {
		return
	}

	lg := logger.FromContext(g.db.context).WithTag("MYSQL")

	if len(v) == 3 && (v[0] == "error" || v[0] == "log") {
		if e, isErr := v[2].(error); isErr {
			// 忽略 context.Canceled 和 TCP取消连接
			errMsg := e.Error()
			if !errors.Is(e, context.Canceled) && !strings.Contains(errMsg, "dial tcp: operation was canceled") {
				lg = lg.WithTag("MYSQL_ERR")
			}
			lg.WithField("file", v[1]).Errorf(errMsg)
			return
		}
	}

	if len(v) == 6 && v[0] == "sql" {
		if g.detailMode {
			writer := logger.Writer()
			values := gorm.LogFormatter(v...)
			_, _ = fmt.Fprint(writer, "\n")
			_, _ = fmt.Fprint(writer, values...)
			_, _ = fmt.Fprint(writer, "\n\n")
		}
		lg.WithFields(map[string]interface{}{
			"file":         v[1],
			"duration":     v[2].(time.Duration).String(),
			"sql":          strings.TrimSpace(v[3].(string)),
			"row_affected": v[5],
		}).Infof("")
		return
	}

	lg.Info(fmt.Sprint(v...))
}
