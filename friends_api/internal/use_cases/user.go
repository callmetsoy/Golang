package useCases

import (
	"github.com/callmetsoy/user_api/internal/domain"
	sqlStorage "github.com/callmetsoy/user_api/internal/storage/sql"
	"github.com/google/uuid"
)

type userRepo interface {
	GetPaginatedUsers(
		page int, pageSize int, filters map[string]interface{}, orderBy string, orderDir string,
	) (domain.PaginatedUsers, error)
	GetCommonFriends(userID1, userID2 uuid.UUID) ([]domain.User, error)
}

type UserUseCase struct {
	userRepo userRepo
}

func NewUserUseCase() *UserUseCase {
	return &UserUseCase{
		userRepo: sqlStorage.GetUserRepo(),
	}
}

func (u *UserUseCase) GetUsers(
	page int, pageSize int, filters map[string]interface{}, orderBy string, orderDir string,
) (domain.PaginatedUsers, error) {
	return u.userRepo.GetPaginatedUsers(page, pageSize, filters, orderBy, orderDir)
}

func (u *UserUseCase) GetCommonFriends(userID1, userID2 uuid.UUID) ([]domain.User, error) {
	return u.userRepo.GetCommonFriends(userID1, userID2)
}
