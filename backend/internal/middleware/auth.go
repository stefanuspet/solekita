package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stefanuspet/solekita/backend/internal/model"
)

// UserContextKey adalah key untuk menyimpan UserClaims di Gin context.
const UserContextKey = "user"

// JWTClaims adalah payload JWT access token.
type JWTClaims struct {
	UserID      uuid.UUID        `json:"user_id"`
	OutletID    uuid.UUID        `json:"outlet_id"`
	OutletCode  string           `json:"outlet_code"`
	Name        string           `json:"name"`
	Phone       string           `json:"phone"`
	IsOwner     bool             `json:"is_owner"`
	Permissions []model.Permission `json:"permissions"`
	jwt.RegisteredClaims
}

// Auth memvalidasi JWT dari header Authorization dan inject UserClaims ke context.
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token tidak ditemukan",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Format token tidak valid",
			})
			return
		}

		tokenString := parts[1]

		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token tidak valid atau sudah expired",
			})
			return
		}

		userClaims := &model.UserClaims{
			ID:          claims.UserID,
			OutletID:    claims.OutletID,
			OutletCode:  claims.OutletCode,
			Name:        claims.Name,
			Phone:       claims.Phone,
			IsOwner:     claims.IsOwner,
			Permissions: claims.Permissions,
		}

		c.Set(UserContextKey, userClaims)
		c.Next()
	}
}

// GetUserFromContext mengambil UserClaims dari Gin context.
// Dipanggil dari handler setelah Auth middleware berjalan.
func GetUserFromContext(c *gin.Context) *model.UserClaims {
	val, _ := c.Get(UserContextKey)
	return val.(*model.UserClaims)
}
