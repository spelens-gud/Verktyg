package gin_middles

import (
	"github.com/gin-gonic/gin"

	"github.com/spelens-gud/Verktyg.git/kits/kcontext"
)

func ExtractMetadata() gin.HandlerFunc {
	return func(c *gin.Context) {
		*c.Request = *c.Request.WithContext(kcontext.SetMetadata(c.Request.Context(), kcontext.Metadata(c.Request.Header)))
	}
}
