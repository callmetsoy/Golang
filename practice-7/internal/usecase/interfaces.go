package usecase

import (
    "practice-7/internal/entity"
    "github.com/google/uuid"
)

type UserInterface interface {
    RegisterUser(user *entity.User) (*entity.User, string, error)
    LoginUser(user *entity.LoginUserDTO) (string, error)
    GetUserByID(userID uuid.UUID) (*entity.User, error)
    PromoteUser(userID uuid.UUID) error
    CountUsers() (int64, error)  // Add this
}