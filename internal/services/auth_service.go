package services

import (
	"errors"
	"time"

	"event-booking/internal/dto"
	"event-booking/internal/models"
	"event-booking/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	users     *repositories.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthService(users *repositories.UserRepository, jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *AuthService) Register(input dto.RegisterRequest) (*dto.AuthResponse, error) {
	_, err := s.users.FindByEmail(input.Email)
	if err == nil {
		return nil, ErrEmailAlreadyUsed
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         input.Role,
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.buildAuthResponse(user)
}

func (s *AuthService) Login(input dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.users.FindByEmail(input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildAuthResponse(user)
}

func (s *AuthService) ParseToken(tokenString string) (dto.UserPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	idValue, ok := claims["sub"].(float64)
	if !ok {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	role, ok := claims["role"].(string)
	if !ok {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	name, ok := claims["name"].(string)
	if !ok {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	email, ok := claims["email"].(string)
	if !ok {
		return dto.UserPayload{}, ErrInvalidCredentials
	}

	return dto.UserPayload{
		ID:    uint(idValue),
		Name:  name,
		Email: email,
		Role:  role,
	}, nil
}

func (s *AuthService) buildAuthResponse(user *models.User) (*dto.AuthResponse, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"role":  user.Role,
		"name":  user.Name,
		"email": user.Email,
		"exp":   expiresAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: tokenString,
		User: dto.UserPayload{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}
