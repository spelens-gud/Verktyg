package memcached

import (
	"git.bestfulfill.tech/devops/go-core/interfaces/istore"
)

func init() {
	istore.RegisteredStore["memcached"] = New
}
