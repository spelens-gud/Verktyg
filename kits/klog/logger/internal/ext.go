package internal

import (
	"context"

	"git.bestfulfill.tech/devops/go-core/interfaces/ilog"
	"git.bestfulfill.tech/devops/go-core/internal/incontext"
)

const keyExtContextPatch = incontext.Key("logger.ext_context_patch")

func ExtPatchFromContext(ctx context.Context) []ilog.LoggerPatch {
	ret, _ := keyExtContextPatch.Value(ctx).([]ilog.LoggerPatch)
	return ret
}

func ExtPatchWithContext(ctx context.Context, p ...ilog.LoggerPatch) context.Context {
	return keyExtContextPatch.WithValue(ctx, p)
}
