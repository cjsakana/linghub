package service

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"linghub/internal/jwtutil"
	"linghub/internal/model"
)

// AuthService handles user registration and login.
type AuthService struct {
	db *gorm.DB
	tm *jwtutil.TokenManager
}

// NewAuthService creates a new authentication service.
func NewAuthService(db *gorm.DB, tm *jwtutil.TokenManager) *AuthService {
	return &AuthService{
		db: db,
		tm: tm,
	}
}

// RegisterRequest holds registration request data.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse holds registration response data.
type RegisterResponse struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
}

// Register registers a new user.
func (s *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// Check if username already exists
	var existingUser model.User
	if result := s.db.Where("username = ?", req.Username).First(&existingUser); result.Error == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", result.Error)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token
	token, err := s.tm.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &RegisterResponse{
		UserID: user.ID,
		Token:  token,
	}, nil
}

// LoginRequest holds login request data.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse holds login response data.
type LoginResponse struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
}

// Login authenticates a user and returns a token.
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by username
	var user model.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate token
	token, err := s.tm.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		UserID: user.ID,
		Token:  token,
	}, nil
}
