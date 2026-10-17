package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	request "github.com/kainguyen/badminton-hub/be/platform/http/request"
)

func RequireBearer(c *gin.Context) {
	if request.BearerToken(c) == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required."}})
		return
	}
	c.Next()
}
