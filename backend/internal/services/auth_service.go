package services

import (
	"database/sql"
	"dfs-backend/internal/database"
	"dfs-backend/internal/dto"
	"dfs-backend/internal/models"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type AuthService struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthService(db *database.DB, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret}
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
	var userModel models.User
	query := "select id, username, email, role, password_hash from users where email = $1"

	if err := s.db.Get(&userModel, query, user.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userModel.PasswordHash), []byte(user.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   userModel.ID.String(),
		"username": userModel.Username,
		"email":    userModel.Email,
		"role":     userModel.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	query := "select id, username, email, role from users where id = $1"

	if err := s.db.Get(&user, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
