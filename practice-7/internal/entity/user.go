package entity

import (
	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"ID" gorm:"primaryKey"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	Role     string    `json:"role"` // user,admin, etc.
	Verified bool      `json:"verified"`
}
