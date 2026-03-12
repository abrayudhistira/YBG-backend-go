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
	GetByName(name string) (entity.User, error)
	DeleteNotIn(activeIDs []string, role string) (int64, error)
	GetByResetToken(token string) (entity.User, error)
	UpdatePassword(userID, newHashedPassword string) error
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
func (r *userRepo) GetByName(name string) (entity.User, error) {
	var u entity.User
	err := r.db.Where("name = ?", name).First(&u).Error
	return u, err
}

//	func (r *userRepo) Update(u *entity.User) error {
//		return r.db.Model(u).
//			Where("user_id = ?", u.UserID).
//			Omit("PointTotal", "PointHistory", "Password", "Role").
//			Updates(u).Error
//	}
func (r *userRepo) Update(u *entity.User) error {
	// Kita definisikan secara eksplisit kolom apa saja yang BOLEH diupdate.
	// Password dan Role TIDAK dimasukkan di sini demi keamanan.
	return r.db.Model(u).
		Where("user_id = ?", u.UserID).
		Select(
			"Name",
			"Email",
			"ProfilePicture",
			"Birth",
			"Phone",
			"Gender",
			"OTPCode",        // Izinkan simpan OTP
			"ResetToken",     // Izinkan simpan Reset Token
			"TokenExpiredAt", // Izinkan simpan Expiry
		).
		Updates(u).Error
}
func (r *userRepo) Delete(id string) error {
	return r.db.Delete(&entity.User{}, "user_id = ?", id).Error
}
func (r *userRepo) DeleteNotIn(ids []string, role string) (int64, error) {
	// Menghapus user yang rolenya customer tapi ID-nya tidak ada di list ids
	result := r.db.Where("role = ? AND user_id NOT IN ?", role, ids).Delete(&entity.User{})
	return result.RowsAffected, result.Error
}
func (r *userRepo) GetByResetToken(token string) (entity.User, error) {
	var user entity.User
	err := r.db.Where("reset_token = ?", token).First(&user).Error
	return user, err
}

// Tambahkan fungsi Update Khusus Password (opsional tapi bagus)
func (r *userRepo) UpdatePassword(userID, newHashedPassword string) error {
	return r.db.Model(&entity.User{}).Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"password":         newHashedPassword,
			"reset_token":      "",
			"otp_code":         "",
			"token_expired_at": nil,
		}).Error
}
