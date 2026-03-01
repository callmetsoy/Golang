package app

import (
	"user_api/internal/repository/_postgres/users"
	"user_api/pkg/modules"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo *users.Repository
}

func NewUserUsecase(repo *users.Repository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) CreateUser(user modules.User) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = string(hashedPassword)

	return u.repo.Create(user)
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetByID(id)
}

func (u *UserUsecase) GetAllUsers() ([]modules.User, error) {
	return u.repo.GetAll()
}

func (u *UserUsecase) UpdateUser(user modules.User) error {
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err == nil {
			user.Password = string(hashedPassword)
		}
	}
	return u.repo.Update(user)
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.Delete(id)
}
