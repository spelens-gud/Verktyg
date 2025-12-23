package memcached

import (
	"github.com/spelens-gud/Verktyg/interfaces/istore"
)

func init() {
	istore.RegisteredStore["memcached"] = New
}
