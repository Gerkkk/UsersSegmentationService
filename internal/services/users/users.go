package users

import (
	"log/slog"
)

type Users struct {
	log  *slog.Logger
	repo UsersRepository
}

type UsersRepository interface {
	CreateUser(id string) (string, error)
	DeleteUser(id string) (string, error)
}

func NewUsers(log *slog.Logger, repo UsersRepository) *Users {
	return &Users{log: log, repo: repo}
}

func (u *Users) CreateUser(id string) (string, error) {
	u.repo.CreateUser(id)
	panic("user service not implemented")
}

func (u *Users) DeleteUser(id string) (string, error) {
	u.repo.DeleteUser(id)
	panic("user service not implemented")
}
