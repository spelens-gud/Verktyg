package structgraph

import (
	"github.com/Just-maple/structgraph"

	"git.bestfulfill.tech/devops/go-core/kits/kstruct/structgraphx"
)

// Deprecated: Use structgraphx.GenStructGraph
func GenStructGraph(in interface{}, filename string, opts ...structgraph.Option) {
	structgraphx.GenStructGraph(in, filename, opts...)
}
