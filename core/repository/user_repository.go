package repository

import (
	"ybg-backend-go/core/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u *entity.User) error
	GetAll() ([]entity.User, error)
	GetByID(id uuid.UUID) (entity.User, error)
	Update(u *entity.User) error
	Delete(id uuid.UUID) error
	GetByEmail(email string) (entity.User, error)
}

type userRepo struct {
	db *gorm.DB
}
func (r *userRepo) GetByEmail(email string) (entity.User, error) {
    var user entity.User
    err := r.db.Where("email = ?", email).First(&user).Error
    return user, err
}
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(u *entity.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) GetAll() ([]entity.User, error) {
	var users []entity.User
	// Kita ikut sertakan PointTotal agar di list user terlihat poinnnya
	err := r.db.Preload("PointTotal").Find(&users).Error
	return users, err
}

func (r *userRepo) GetByID(id uuid.UUID) (entity.User, error) {
	var user entity.User
	// Preload PointTotal dan History untuk detail profil
	err := r.db.Preload("PointTotal").Preload("PointHistory").First(&user, "user_id = ?", id).Error
	return user, err
}

func (r *userRepo) Update(u *entity.User) error {
	// Omit PointTotal agar update user tidak menimpa data poin secara tidak sengaja
	return r.db.Model(u).Omit("PointTotal", "PointHistory").Updates(u).Error
}

func (r *userRepo) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.User{}, "user_id = ?", id).Error
}
