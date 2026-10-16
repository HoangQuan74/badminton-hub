package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewTokenManager(secret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}
}

func (m *TokenManager) NewAccessToken(userID string) (string, time.Time, error) {
	now := m.now().UTC()
	expiresAt := now.Add(m.accessTTL)
	claims := jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(expiresAt)}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return token, expiresAt, nil
}

func (m *TokenManager) ParseAccessToken(raw string) (string, error) {
	claims := new(jwt.RegisteredClaims)
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid || claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}

func (m *TokenManager) NewRefreshToken() (raw string, hash []byte, expiresAt time.Time, err error) {
	buffer := make([]byte, 32)
	if _, err = rand.Read(buffer); err != nil {
		return "", nil, time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(buffer)
	hash = HashRefreshToken(raw)
	return raw, hash, m.now().UTC().Add(m.refreshTTL), nil
}

func HashRefreshToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
