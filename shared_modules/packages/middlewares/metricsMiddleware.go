package middlewares

import (
	"fmt"

	"github.com/coni63/my_app/shared_modules/packages/initializers"

	"github.com/gin-gonic/gin"
)

// Prometheus middleware for Gin
func PrometheusStatusCodeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		initializers.CountRequests.Inc() // Increment the counter for each request

		// Process request
		c.Next()

		// After request is handled, get the status code
		statusCode := c.Writer.Status()
		initializers.CountStatusCodes.WithLabelValues(fmt.Sprintf("%d", statusCode)).Inc()
	}
}
