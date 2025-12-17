package structgraph

import (
	"github.com/Just-maple/structgraph"

	"github.com/spelens-gud/Verktyg.git/kits/kstruct/structgraphx"
)

// Deprecated: Use structgraphx.GenStructGraph
func GenStructGraph(in interface{}, filename string, opts ...structgraph.Option) {
	structgraphx.GenStructGraph(in, filename, opts...)
}
