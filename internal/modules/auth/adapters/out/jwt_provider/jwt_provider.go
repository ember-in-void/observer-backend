package jwt_provider

import (
	"context"
	"errors"
	"time"

	"steam-observer/internal/modules/auth/ports/out_ports"
	"steam-observer/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type jwtProvider struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTProvider - создаёт провайдер JWT токенов
func NewJWTProvider(cfg config.JWTConfig) out_ports.TokenProvider {
	return &jwtProvider{
		secret: []byte(cfg.Secret),
		ttl:    cfg.TTL,
	}
}

// Claims - кастомные claims для JWT
type Claims struct {
	UserID string  `json:"user_id"`
	Email  *string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAccessToken - генерирует JWT access token
func (p *jwtProvider) GenerateAccessToken(ctx context.Context, userID string, email *string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(p.ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "steam-observer",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(p.secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken - валидирует JWT токен и возвращает claims
func (p *jwtProvider) ValidateToken(ctx context.Context, tokenString string) (*out_ports.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверка алгоритма подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &out_ports.TokenClaims{
			UserID: claims.UserID,
			Email:  claims.Email,
		}, nil
	}

	return nil, errors.New("invalid token")
}

// ParseAccessToken - парсит токен и возвращает userID + email
// Это wrapper вокруг ValidateToken для удобства middleware
func (p *jwtProvider) ParseAccessToken(ctx context.Context, tokenString string) (string, *string, error) {
	// Переиспользуем ValidateToken
	claims, err := p.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", nil, err
	}

	return claims.UserID, claims.Email, nil
}
