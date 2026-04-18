package runtime

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := extractBearerToken(authHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		u, err := s.loadAuthedUser(token, nowStr())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			c.Abort()
			return
		}

		c.Set(string(userKey), u)
		c.Next()
	}
}

func getAuthedUser(c *gin.Context) authedUser {
	val, exists := c.Get(string(userKey))
	if !exists {
		return authedUser{}
	}
	return val.(authedUser)
}

func extractBearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(strings.ToLower(authHeader), strings.ToLower(prefix)) {
		return ""
	}
	return strings.TrimSpace(authHeader[len(prefix):])
}
