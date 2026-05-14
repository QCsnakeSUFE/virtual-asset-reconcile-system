package middleware

import (
	"fmt"

	"virtual-asset-reconcile-system/pkg/idgen"

	"github.com/gin-gonic/gin"
)

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = fmt.Sprintf("%d", idgen.NextID())
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)

		tenantID := c.GetHeader("X-Tenant-Id")
		if tenantID == "" {
			tenantID = "default"
		}
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}
