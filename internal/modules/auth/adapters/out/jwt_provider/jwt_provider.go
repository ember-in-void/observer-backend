package jwt_provider

import (
	"context"
	"fmt"
	"time"

	"steam-observer/internal/modules/auth/ports/out_ports"
	"steam-observer/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTProvider(cfg config.JWTConfig) out_ports.TokenProvider {
	return &JWTProvider{
		secret: []byte(cfg.Secret),
		ttl:    cfg.TTL,
	}
}

type Claims struct {
	UserID string
	Email  *string
	jwt.RegisteredClaims
}

func (p *JWTProvider) GenerateAccessToken(ctx context.Context, userID string, email *string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.secret)
}

func (p *JWTProvider) ParseAccessToken(ctx context.Context, tokenStr string) (string, *string, error) {
	parsed, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
	})
	if err != nil || !parsed.Valid {
		return "", nil, fmt.Errorf("invalid token")
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return "", nil, fmt.Errorf("invalid claims type")
	}

	return claims.UserID, claims.Email, nil
}
