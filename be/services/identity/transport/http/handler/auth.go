package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kainguyen/badminton-hub/be/services/identity/adapters/postgres"
	identityconfig "github.com/kainguyen/badminton-hub/be/services/identity/config"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/apperror"
	apiMiddleware "github.com/kainguyen/badminton-hub/be/services/identity/transport/http/middleware"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/models/request"
	"github.com/kainguyen/badminton-hub/be/services/identity/transport/http/models/response"
)

type AuthDependencies struct {
	Config identityconfig.Config
	Pool   *pgxpool.Pool
	Logger *slog.Logger
}

type AuthHandler struct {
	service *auth.Service
}

// RegisterAuthRoutes initializes the Auth domain and registers its HTTP routes.
func RegisterAuthRoutes(engine *gin.Engine, dependencies AuthDependencies) {
	repository := postgres.NewAuthRepository(dependencies.Pool)
	tokens := auth.NewTokenManager(dependencies.Config.JWTSecret, dependencies.Config.AccessTokenTTL, dependencies.Config.RefreshTokenTTL)
	service := auth.NewService(repository, tokens)
	handler := &AuthHandler{service: service}

	v1 := engine.Group("/api/v1")
	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", apiMiddleware.HandleJSON(dependencies.Logger, handler.Register))
	authRoutes.POST("/login", apiMiddleware.HandleJSON(dependencies.Logger, handler.Login))
	authRoutes.POST("/refresh", apiMiddleware.HandleJSON(dependencies.Logger, handler.Refresh))
	authRoutes.POST("/logout", apiMiddleware.HandleJSON(dependencies.Logger, handler.Logout))
	v1.GET("/me", apiMiddleware.RequireAuth(tokens), apiMiddleware.Handle(dependencies.Logger, handler.Me))
}

func (h *AuthHandler) Register(c *gin.Context, input request.RegisterRequest) error {
	created, pair, err := h.service.Register(c.Request.Context(), auth.RegisterInput{Email: input.Email, Password: input.Password, FullName: input.FullName, Phone: input.Phone})
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, response.AuthResponse{User: response.NewUser(created), Tokens: response.NewToken(pair)})
	return nil
}

func (h *AuthHandler) Login(c *gin.Context, input request.LoginRequest) error {
	found, pair, err := h.service.Login(c.Request.Context(), auth.LoginInput{Email: input.Email, Password: input.Password})
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, response.AuthResponse{User: response.NewUser(found), Tokens: response.NewToken(pair)})
	return nil
}

func (h *AuthHandler) Refresh(c *gin.Context, input request.RefreshTokenRequest) error {
	pair, err := h.service.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, response.RefreshResponse{Tokens: response.NewToken(pair)})
	return nil
}

func (h *AuthHandler) Logout(c *gin.Context, input request.RefreshTokenRequest) error {
	if err := h.service.Logout(c.Request.Context(), input.RefreshToken); err != nil {
		return err
	}
	c.Status(http.StatusNoContent)
	return nil
}

func (h *AuthHandler) Me(c *gin.Context) error {
	userID, ok := apiMiddleware.UserID(c)
	if !ok {
		return apperror.Unauthorized()
	}
	found, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, response.MeResponse{User: response.NewUser(found)})
	return nil
}
