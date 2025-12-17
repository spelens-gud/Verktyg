package memcached

import (
	"github.com/spelens-gud/Verktyg.git/interfaces/istore"
)

func init() {
	istore.RegisteredStore["memcached"] = New
}
