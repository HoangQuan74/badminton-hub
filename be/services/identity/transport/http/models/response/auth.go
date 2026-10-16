package response

import (
	"time"

	"github.com/kainguyen/badminton-hub/be/services/identity/domain/auth"
	"github.com/kainguyen/badminton-hub/be/services/identity/domain/user"
)

type UserResponse struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	FullName  string      `json:"full_name"`
	Phone     string      `json:"phone"`
	Status    user.Status `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type TokenResponse struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

type RefreshResponse struct {
	Tokens TokenResponse `json:"tokens"`
}

type MeResponse struct {
	User UserResponse `json:"user"`
}

func NewUser(value user.User) UserResponse {
	return UserResponse{ID: value.ID, Email: value.Email, FullName: value.FullName, Phone: value.Phone, Status: value.Status, CreatedAt: value.CreatedAt}
}

func NewToken(pair auth.TokenPair) TokenResponse {
	return TokenResponse{AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, TokenType: "Bearer", AccessExpiresAt: pair.AccessExpiresAt, RefreshExpiresAt: pair.RefreshExpiresAt}
}
