package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ryanpzr/shopping-cart-api/pkg/apperrors"
	jwtutil "github.com/ryanpzr/shopping-cart-api/pkg/jwt"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			apperrors.HandleError(c, apperrors.NewUnauthorized("missing or invalid authorization header"))
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtutil.Parse(tokenStr)
		if err != nil {
			apperrors.HandleError(c, err)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		userRole, _ := role.(string)
		for _, r := range roles {
			if r == userRole {
				c.Next()
				return
			}
		}
		apperrors.HandleError(c, apperrors.NewForbidden("insufficient permissions"))
		c.Abort()
	}
}
