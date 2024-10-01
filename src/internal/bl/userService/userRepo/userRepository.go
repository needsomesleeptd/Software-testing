package repository

import "annotater/internal/models"

type IUserRepository interface {
	GetUserByLogin(login string) (*models.User, error)
	UpdateUserByLogin(login string, user *models.User) error
	CreateUser(user *models.User) error
	GetAllUsers() ([]models.User, error)
}
