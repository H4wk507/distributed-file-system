package services

import (
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/models"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) RegisterUser(user *dto.RegisterUserRequest) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	passwordHash := string(bytes)

	userModel := &models.User{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: passwordHash,
		Role:         models.UserRoleUser,
	}

	query := `
		insert into users (username, email, password_hash, role)
		values ($1, $2, $3, $4)
	`
	if _, err = s.db.Exec(query, userModel.Username, userModel.Email, userModel.PasswordHash, userModel.Role); err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	return nil
}

func (s *AuthService) LoginUser(user *dto.LoginUserRequest) (string, error) {
	var passwordHash string
	query := "select password_hash from users where email = $1"

	if err := s.db.Get(&passwordHash, query, user.Email); err != nil {
		return "", fmt.Errorf("failed to get password hash: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)); err != nil {
		return "", fmt.Errorf("invalid password: %w", err)
	}

	// TODO: issue JWT token
	return "", nil
}
