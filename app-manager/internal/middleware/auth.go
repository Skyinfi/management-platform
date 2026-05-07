package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AuthClaims struct {
	Subject   string
	Name      string
	Role      string
	ExpiresAt time.Time
}

type JWTValidator interface {
	Validate(token string) (AuthClaims, error)
}

type authContextKey struct{}

func Auth(validator JWTValidator) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shouldSkipAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if token == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			claims, err := validator.Validate(token)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), authContextKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func shouldSkipAuth(path string) bool {
	switch path {
	case "/api/health", "/api/auth/login", "/api/auth/me":
		return true
	default:
		return false
	}
}

func ClaimsFromContext(ctx context.Context) (AuthClaims, bool) {
	claims, ok := ctx.Value(authContextKey{}).(AuthClaims)
	return claims, ok
}

type JWTService struct {
	secret []byte
}

type jwtPayload struct {
	Subject   string `json:"sub"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) Sign(subject, name, role string, expiresAt time.Time) (string, error) {
	payload := jwtPayload{Subject: subject, Name: name, Role: role, ExpiresAt: expiresAt.Unix()}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(body)
	signature := s.sign(encodedPayload)
	return encodedPayload + "." + signature, nil
}

func (s *JWTService) Validate(token string) (AuthClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return AuthClaims{}, errors.New("invalid token")
	}

	if !hmac.Equal([]byte(parts[1]), []byte(s.sign(parts[0]))) {
		return AuthClaims{}, errors.New("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return AuthClaims{}, err
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return AuthClaims{}, err
	}

	if time.Now().After(time.Unix(payload.ExpiresAt, 0)) {
		return AuthClaims{}, errors.New("token expired")
	}

	return AuthClaims{Subject: payload.Subject, Name: payload.Name, Role: payload.Role, ExpiresAt: time.Unix(payload.ExpiresAt, 0)}, nil
}

func (s *JWTService) sign(payload string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

var _ JWTValidator = (*JWTService)(nil)

func (s *JWTService) String() string {
	return fmt.Sprintf("jwt-service(%d bytes)", len(s.secret))
}
