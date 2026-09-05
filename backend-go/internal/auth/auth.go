package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenTTL = 7 * 24 * time.Hour

type Claims struct {
	UserID    string
	SessionID string
	AdminID   string
	Type      string
}

func SignToken(secret string, claims Claims) (string, error) {
	now := time.Now()
	mc := jwt.MapClaims{
		"type": claims.Type,
		"iat":  now.Unix(),
		"exp":  now.Add(TokenTTL).Unix(),
	}
	if claims.UserID != "" {
		mc["userId"] = claims.UserID
	}
	if claims.SessionID != "" {
		mc["sessionId"] = claims.SessionID
	}
	if claims.AdminID != "" {
		mc["adminId"] = claims.AdminID
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, mc).SignedString([]byte(secret))
}

func VerifyToken(secret, token string) (*Claims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	out := &Claims{Type: str(mc, "type")}
	out.UserID = str(mc, "userId")
	out.SessionID = str(mc, "sessionId")
	out.AdminID = str(mc, "adminId")
	return out, nil
}

func str(mc jwt.MapClaims, key string) string {
	if v, ok := mc[key].(string); ok {
		return v
	}
	return ""
}

// ParseBearerToken mirrors the Node helper: empty when header missing/malformed.
func ParseBearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}
