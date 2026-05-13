package repository

import (
	"gorm.io/gorm"

	"order-system/internal/pkg/db"
)

type UserRepository interface {
	FindByID(id uint) (*db.User, error)
	FindByUsername(username string) (*db.User, error)
	Create(user *db.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint) (*db.User, error) {
	var user db.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByUsername(username string) (*db.User, error) {
	var user db.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *db.User) error {
	return r.db.Create(user).Error
}
