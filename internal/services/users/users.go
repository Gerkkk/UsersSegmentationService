package users

import (
	"log/slog"
	"main/internal/domain/models"
	apperrors "main/internal/errors"
)

// Users - структура сервиса для управления пользователями
type Users struct {
	log   *slog.Logger
	repo  UsersRepository
	cache SegmentationCache
}

type UsersRepository interface {
	CreateUser(user models.User) (int, error)
	DeleteUser(id int) (int, error)
}

type SegmentationCache interface {
	SaveUserSegments(key int, val []models.Segment) error
	TryGetUserSegments(key int) ([]models.Segment, error)
	Invalidate() error
}

func NewUsers(log *slog.Logger, repo UsersRepository, cache SegmentationCache) *Users {
	return &Users{log: log, repo: repo, cache: cache}
}

// CreateUser - Создание пользователя с заданной структорой
func (u *Users) CreateUser(user models.User) (int, error) {
	id, err := u.repo.CreateUser(user)

	if err != nil {
		return -1, apperrors.Convert(u.log, err)
	}

	return id, nil
}

// DeleteUser - удаление пользователя по id
func (u *Users) DeleteUser(id int) (int, error) {
	id, err := u.repo.DeleteUser(id)

	if err != nil {
		return -1, apperrors.Convert(u.log, err)
	}

	err = u.cache.Invalidate()

	if err != nil {
		return -1, apperrors.Convert(u.log, err)
	}

	return id, nil
}
