package middleware

import (
	"net/http"
	"strings"

	"event-booking/internal/dto"
	"event-booking/internal/services"

	"github.com/gin-gonic/gin"
)

const userContextKey = "auth_user"

func Auth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		user, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(userContextKey, user)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := MustUser(c)
		if user.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func MustUser(c *gin.Context) dto.UserPayload {
	value, _ := c.Get(userContextKey)
	user, _ := value.(dto.UserPayload)
	return user
}

func SetUserDetails(c *gin.Context, user dto.UserPayload) {
	c.Set(userContextKey, user)
}
