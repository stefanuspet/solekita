package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

// RequireOwner memastikan hanya owner yang bisa mengakses route.
// Harus dipasang setelah Auth middleware.
func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c)
		if !user.IsOwner {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Hanya owner yang bisa mengakses ini",
			})
			return
		}
		c.Next()
	}
}

// RequirePermission memastikan user memiliki salah satu dari permission yang diberikan.
// Owner bypass semua permission check.
// Harus dipasang setelah Auth middleware.
func RequirePermission(permissions ...model.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetUserFromContext(c)

		if user.IsOwner {
			c.Next()
			return
		}

		for _, p := range permissions {
			if user.HasPermission(p) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Tidak punya akses untuk melakukan tindakan ini",
		})
	}
}
