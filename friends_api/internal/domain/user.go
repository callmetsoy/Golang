package domain

import (
	"time"

	"github.com/google/uuid"
)

// ENUM(M, F)
type Gender string

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Gender    Gender    `json:"gender"`
	BirthDate time.Time `json:"birth_date"`
}

type PaginatedUsers struct {
	Data       []User `json:"data"`
	TotalCount int    `json:"totalCount"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}
