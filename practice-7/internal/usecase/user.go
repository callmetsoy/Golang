package usecase

import (
    "fmt"
    "practice-7/internal/entity"
    "practice-7/internal/usecase/repo"
    "practice-7/utils"
    "github.com/google/uuid"
)

type UserUseCase struct {
    repo *repo.UserRepo
}

func NewUserUseCase(r *repo.UserRepo) *UserUseCase {
    return &UserUseCase{
        repo: r,
    }
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
    count, err := u.repo.CountUsers()
    if err != nil {
        return nil, "", fmt.Errorf("count users: %w", err)
    }
    if count == 0 {
        user.Role = "admin"  // First user is admin
    } else {
        user.Role = "user"
    }
    user.ID = uuid.New()  // <-- Добавьте эту строку здесь
    user, err = u.repo.RegisterUser(user)
    if err != nil {
        return nil, "", fmt.Errorf("register user: %w", err)
    }
    sessionID := uuid.New().String()
    return user, sessionID, nil
}

func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
    userFromRepo, err := u.repo.LoginUser(user)
    if err != nil {
        return "", fmt.Errorf("User From Repo: %w", err)
    }
    if !utils.CheckPassword(userFromRepo.Password, user.Password) {
        return "", fmt.Errorf("Check Password: %w", err)
    }
    token, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
    if err != nil {
        return "", fmt.Errorf("Generate JWT: %w", err)
    }
    return token, nil
}

func (u *UserUseCase) GetUserByID(userID uuid.UUID) (*entity.User, error) {
    user, err := u.repo.GetUserByID(userID)
    if err != nil {
        return nil, fmt.Errorf("get user by ID: %w", err)
    }
    return user, nil
}

func (u *UserUseCase) PromoteUser(userID uuid.UUID) error {
    err := u.repo.PromoteUser(userID)
    if err != nil {
        return fmt.Errorf("promote user: %w", err)
    }
    return nil
}

func (u *UserUseCase) CountUsers() (int64, error) {
    return u.repo.CountUsers()
}