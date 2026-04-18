package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuth memvalidasi header X-Admin-Key untuk akses admin panel.
func AdminAuth(adminSecretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Admin-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Admin key tidak ditemukan",
			})
			return
		}

		if key != adminSecretKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin key tidak valid",
			})
			return
		}

		c.Next()
	}
}
