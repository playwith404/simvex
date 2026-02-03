package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const userIDKey = "user_id"

func AuthMiddleware(redisClient *redis.Client, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(cookieName)
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "로그인이 필요합니다",
			})
			return
		}
		userID, err := redisClient.Get(c.Request.Context(), "session:"+sessionID).Result()
		if err == redis.Nil || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "로그인이 필요합니다",
			})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code":    "SESSION_ERROR",
				"message": "세션 확인에 실패했습니다",
			})
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}
	id, ok := value.(string)
	return id, ok
}
