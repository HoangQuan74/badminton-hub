package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteInvalidRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
}
