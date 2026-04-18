package service

import (
	"Practice-8/repository"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	user := &repository.User{ID: 1, Name: "Test", Email: "test@mail.com"}

	t.Run("user exists", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(user.Email).Return(user, nil)

		err := service.RegisterUser(user, user.Email)
		assert.Error(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(user.Email).Return(nil, errors.New("db error"))

		err := service.RegisterUser(user, user.Email)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().GetByEmail(user.Email).Return(nil, nil)
		mockRepo.EXPECT().CreateUser(user).Return(nil)

		err := service.RegisterUser(user, user.Email)
		assert.NoError(t, err)
	})
}

func TestUpdateUserName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	t.Run("empty name", func(t *testing.T) {
		err := service.UpdateUserName(1, "")
		assert.Error(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByID(1).Return(nil, errors.New("not found"))

		err := service.UpdateUserName(1, "New")
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		user := &repository.User{ID: 1, Name: "Old"}

		mockRepo.EXPECT().GetUserByID(1).Return(user, nil)
		mockRepo.EXPECT().UpdateUser(user).Return(nil)

		err := service.UpdateUserName(1, "New")
		assert.NoError(t, err)
		assert.Equal(t, "New", user.Name)
	})
}

func TestDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)

	t.Run("delete admin", func(t *testing.T) {
		err := service.DeleteUser(1)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().DeleteUser(2).Return(nil)

		err := service.DeleteUser(2)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo.EXPECT().DeleteUser(3).Return(errors.New("db error"))

		err := service.DeleteUser(3)
		assert.Error(t, err)
	})
}
