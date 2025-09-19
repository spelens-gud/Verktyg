package irpc

import "context"

type RpcClient interface {
	Call(ctx context.Context, path string, param interface{}, resp interface{}) (err error)
}
