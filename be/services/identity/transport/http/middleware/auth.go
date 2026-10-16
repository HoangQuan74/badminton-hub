package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
	localization "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/i18n"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/models/response"
)

const userIDKey = "authenticated_user_id"

func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, raw, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(raw) == "" {
			abortUnauthorized(c)
			return
		}
		userID, err := tokens.ParseAccessToken(strings.TrimSpace(raw))
		if err != nil {
			abortUnauthorized(c)
			return
		}
		c.Set(userIDKey, userID)
		c.Next()
	}
}

func UserID(c *gin.Context) (string, bool) {
	value, ok := c.Get(userIDKey)
	userID, valid := value.(string)
	return userID, ok && valid && userID != ""
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorEnvelope{Error: response.ErrorDetail{Code: "UNAUTHORIZED", Message: localization.T(c, "error.unauthorized")}})
}
