package kcontext

import (
	"context"
	"net/http"
)

func SetRequestContext(r *http.Request, ctx context.Context) {
	*r = *r.WithContext(ctx)
}

func SetRequestServiceCode(r *http.Request, code int) {
	SetRequestContext(r, SetServiceCode(r.Context(), code))
}

func SetRequestError(r *http.Request, err error) {
	SetRequestContext(r, context.WithValue(r.Context(), requestErrorKey, err))
}
