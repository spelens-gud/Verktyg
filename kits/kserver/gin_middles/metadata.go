package gin_middles

import (
	"github.com/gin-gonic/gin"

	"git.bestfulfill.tech/devops/go-core/kits/kcontext"
)

func ExtractMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		*c.Request = *c.Request.WithContext(kcontext.SetMetadata(c.Request.Context(), kcontext.Metadata(c.Request.Header)))
	}
}
