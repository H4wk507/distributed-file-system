package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID uuid.UUID `db:"id" json:"id"`

	Username     string   `db:"username" json:"username"`
	Email        string   `db:"email" json:"email"`
	PasswordHash string   `db:"password_hash" json:"password_hash"`
	Role         UserRole `db:"role" json:"role"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
