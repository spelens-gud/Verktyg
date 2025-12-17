package structgraphx

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Just-maple/structgraph"

	"github.com/spelens-gud/Verktyg.git/kits/klog/logger"
)

func GenStructGraph(in interface{}, filename string, opts ...structgraph.Option) {
	if in == nil {
		panic("invalid struct")
	}

	lg := logger.FromBackground()

	if len(filename) == 0 {
		filename = "./design/structure.png"
	}

	_ = os.MkdirAll(filepath.Dir(filename), 0775)

	if _, e := exec.LookPath("dot"); e == nil {
		if err := os.WriteFile("design/structure.dot", []byte(structgraph.Draw(in)), 0664); err != nil {
			return
		}
		defer func() {
			_ = os.Remove("design/structure.dot")
		}()
		if err := exec.Command("/bin/sh", "-c", "dot -T png design/structure.dot -o "+filename).Run(); err != nil {
			logger.FromContext(context.Background()).Errorf("gen structure error: %v", err)
		}
		return
	}

	ret := structgraph.Draw(in, opts...)
	if len(ret) == 0 {
		return
	}

	if err := structgraph.GenPngFromQuickChartApi(ret, filename); err != nil {
		lg.Errorf("gen structure graph error: %v", err)
		return
	}

	lg.Infof("gen structure graph [ %s ] success", filename)
}
