package service

import (
	"fmt"
	"time"

	"github.com/Skyinfi/app-manager/internal/config"
	"github.com/Skyinfi/app-manager/internal/middleware"
	"github.com/Skyinfi/app-manager/internal/model"
)

type AuthService struct {
	cfg config.Config
	jwt *middleware.JWTService
}

func NewAuth(cfg config.Config) *AuthService {
	return &AuthService{cfg: cfg, jwt: middleware.NewJWTService(cfg.JWTSecret)}
}

func (s *AuthService) Login(req model.LoginRequest) (model.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return model.LoginResponse{}, fmt.Errorf("用户名或密码不能为空")
	}

	user := model.LoginUser{ID: "1", Name: "管理员", Role: "admin"}
	expiresAt := time.Now().Add(24 * time.Hour)
	token, err := s.jwt.Sign(user.ID, user.Name, user.Role, expiresAt)
	if err != nil {
		return model.LoginResponse{}, err
	}

	return model.LoginResponse{Token: token, ExpiresAt: expiresAt.Format(time.RFC3339), User: user}, nil
}

func (s *AuthService) Me(token string) (model.MeResponse, error) {
	claims, err := s.jwt.Validate(token)
	if err != nil {
		return model.MeResponse{}, err
	}

	return model.MeResponse{
		Token:     token,
		ExpiresAt: claims.ExpiresAt.Format(time.RFC3339),
		User:      model.LoginUser{ID: claims.Subject, Name: claims.Name, Role: claims.Role},
	}, nil
}
