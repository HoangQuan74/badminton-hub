package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required."}})
}
