package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AbortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "UNAUTHORIZED", "message": "Authentication required."}})
}

func WriteImageRequired(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "A JPEG or PNG file is required (maximum 2 MiB)."}})
}

func WriteImageInvalid(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": "Invalid image or file exceeds 2 MiB."}})
}
