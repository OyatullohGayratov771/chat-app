// internal/utils/jwt.go
package utils

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTProvider struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
	redis         *redis.Client
}

func NewJWTProvider(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration, r *redis.Client) *JWTProvider {
	return &JWTProvider{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
		redis:         r,
	}
}

func (p *JWTProvider) GenerateTokens(userID string) (string, string, error) {
	now := time.Now()
	accessClaims := jwt.MapClaims{
		"userID": userID,
		"exp":     now.Add(p.accessTTL).Unix(),
		"iat":     now.Unix(),
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := access.SignedString(p.accessSecret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"userID": userID,
		"exp":     now.Add(p.refreshTTL).Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.New().String(), // unique id for revocation
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refresh.SignedString(p.refreshSecret)
	if err != nil {
		return "", "", err
	}

	// save refresh token jti -> userID mapping in Redis (or store whole token)
	// TTL equal to refreshTTL for automatic expiration
	jti, _ := refreshClaims["jti"].(string)
	ctx := context.Background()
	if err := p.redis.Set(ctx, "refresh:"+jti, userID, p.refreshTTL).Err(); err != nil {
		return "", "", err
	}

	// Optionally map token string -> jti for quick revoke by token string
	return accessStr, refreshStr, nil
}

func (p *JWTProvider) ValidateAccessToken(tokenStr string) (userID string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	userID, ok = claims["userID"].(string)
	if !ok {
		return "", errors.New("invalid userID")
	}
	return userID, nil
}

func (p *JWTProvider) ValidateRefreshToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return p.refreshSecret, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	userID, ok := claims["userID"].(string)
	if !ok {
		return "", errors.New("invalid userID")
	}
	jti, ok := claims["jti"].(string)
	if !ok {
		return "", errors.New("jti missing")
	}

	// check redis
	ctx := context.Background()
	val, err := p.redis.Get(ctx, "refresh:"+jti).Result()
	if err == redis.Nil {
		return "", errors.New("revoked or not found")
	}
	if err != nil {
		return "", err
	}
	if val != userID {
		return "", errors.New("token mismatch")
	}
	return userID, nil
}

func (p *JWTProvider) RevokeRefreshToken(tokenStr string) error {
	// parse jti and delete key
	token, _ := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return p.refreshSecret, nil })
	if token == nil {
		return errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	jti := claims["jti"].(string)
	return p.redis.Del(context.Background(), "refresh:"+jti).Err()
}
