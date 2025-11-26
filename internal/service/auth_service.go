package service

import "github.com/alifdwt/techtest-indico-be/internal/dto"

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	return &dto.LoginResponse{
		Token: "iniadalahtokenbohongan",
	}, nil
}
