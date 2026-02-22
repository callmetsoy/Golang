package repository

import (
	"user_api/internal/repository/_postgres"
	"user_api/internal/repository/_postgres/users"
	"user_api/pkg/modules"
)

type UserRepository interface {
	Create(u modules.User) (int, error)
	GetByID(id int) (*modules.User, error)
	GetAll() ([]modules.User, error)
	Update(u modules.User) error
	Delete(id int) (int64, error)
}

type Repositories struct {
	UserRepository UserRepository
}

func NewRepositories(db *_postgres.PGXDialect) *Repositories { // Замени на PGXDialect
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}
