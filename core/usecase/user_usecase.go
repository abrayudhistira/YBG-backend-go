package usecase

import (
	"errors"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/pkg/utils"

	"github.com/google/uuid"
)

type UserUsecase interface {
	RegisterUser(u *entity.User) error
	FetchAllUsers() ([]entity.User, error)
	GetUserProfile(id uuid.UUID) (entity.User, error)
	UpdateProfile(u *entity.User) error
	RemoveUser(id uuid.UUID) error
	Login(email, password string) (entity.User, error)
}

type userUC struct {
	repo      repository.UserRepository
	pointRepo repository.PointRepository
}

func NewUserUsecase(repo repository.UserRepository, pointRepo repository.PointRepository) UserUsecase {
	return &userUC{
		repo:      repo,
		pointRepo: pointRepo,
	}
}

// func (u *userUC) RegisterUser(user *entity.User) error {
// 	if user.UserID == uuid.Nil {
// 		user.UserID = uuid.New()
// 	}

// 	// Hashing Password sebelum simpan ke DB
// 	hashedPassword, err := utils.HashPassword(user.Password)
// 	if err != nil {
// 		return err
// 	}
// 	user.Password = hashedPassword

// 	return u.repo.Create(user)
// }

func (u *userUC) RegisterUser(user *entity.User) error {
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}

	// 1. Hashing Password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	// 2. Simpan User ke DB
	err = u.repo.Create(user)
	if err != nil {
		return err
	}

	// 3. INISIALISASI POINT_TOTAL (WAJIB!)
	// Supaya user baru langsung punya 'dompet' poin
	pointTotal := entity.PointTotal{
		UserID:    user.UserID,
		Total:     0,
		Tier:      "friend", // Default tier sesuai model kamu
		CreatedAt: time.Now(),
	}

	// Asumsikan kamu sudah punya method CreatePointTotal di PointRepository
	return u.pointRepo.CreatePointTotal(&pointTotal)
}

func (u *userUC) FetchAllUsers() ([]entity.User, error) {
	return u.repo.GetAll()
}

func (u *userUC) GetUserProfile(id uuid.UUID) (entity.User, error) {
	return u.repo.GetByID(id)
}

func (u *userUC) UpdateProfile(user *entity.User) error {
	return u.repo.Update(user)
}

func (u *userUC) RemoveUser(id uuid.UUID) error {
	return u.repo.Delete(id)
}

func (u *userUC) Login(email, password string) (entity.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return entity.User{}, err // user not found or db error
	}
	// Compare hashed password
	if !utils.CheckPasswordHash(password, user.Password) {
		return entity.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
