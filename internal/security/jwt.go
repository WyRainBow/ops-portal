package security

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func secret() []byte {
	if v := os.Getenv("OPS_PORTAL_JWT_SECRET"); v != "" {
		return []byte(v)
	}
	// dev default; must be overridden in production.
	return []byte("change-me-in-production")
}

func expireHours() time.Duration {
	h := 168
	if v := os.Getenv("OPS_PORTAL_JWT_EXPIRE_HOURS"); v != "" {
		if x, err := strconv.Atoi(v); err == nil && x > 0 {
			h = x
		}
	}
	return time.Duration(h) * time.Hour
}

func CreateToken(userID int64, username string, role string) (string, error) {
	now := time.Now().UTC()
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expireHours())),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(secret())
}

func ParseToken(token string) (*Claims, error) {
	if token == "" {
		return nil, errors.New("empty token")
	}
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return secret(), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return c, nil
}

