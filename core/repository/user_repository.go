package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *entity.User) error
	GetAll() ([]entity.User, error)
	GetByID(id string) (entity.User, error) // Ganti ke string
	Update(u *entity.User) error
	Delete(id string) error // Ganti ke string
	GetByEmail(email string) (entity.User, error)
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) GetAll() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Preload("PointTotal").Find(&users).Error
	return users, err
}

func (r *userRepo) GetByID(id string) (entity.User, error) {
	var user entity.User
	// GORM akan otomatis menangani string ID di sini
	err := r.db.Preload("PointTotal").Preload("PointHistory").First(&user, "user_id = ?", id).Error
	return user, err
}

func (r *userRepo) GetByEmail(email string) (entity.User, error) {
	var user entity.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *userRepo) Update(u *entity.User) error {
	return r.db.Model(u).
		Where("user_id = ?", u.UserID).
		Omit("PointTotal", "PointHistory", "Password", "Role").
		Updates(u).Error
}

func (r *userRepo) Delete(id string) error {
	return r.db.Delete(&entity.User{}, "user_id = ?", id).Error
}
