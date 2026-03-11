package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type PointRepository interface {
	CreateHistory(h *entity.PointHistory) error
	UpdateTotal(uid string, addedPoint int) error                 // Ganti ke string
	GetHistoryByUserID(uid string) ([]entity.PointHistory, error) // Ganti ke string
	CreatePointTotal(pt *entity.PointTotal) error
	GetAllTotalsWithUser() ([]entity.PointTotal, error)
}

type pointRepo struct {
	db *gorm.DB
}

func NewPointRepository(db *gorm.DB) PointRepository { return &pointRepo{db: db} }

func (r *pointRepo) CreateHistory(h *entity.PointHistory) error {
	return r.db.Create(h).Error
}

func (r *pointRepo) UpdateTotal(uid string, addedPoint int) error {
	return r.db.Model(&entity.PointTotal{}).
		Where("user_id = ?", uid).
		Update("total", gorm.Expr("total + ?", addedPoint)).Error
}

func (r *pointRepo) GetHistoryByUserID(uid string) ([]entity.PointHistory, error) {
	var history []entity.PointHistory
	err := r.db.Where("user_id = ?", uid).Order("created_at desc").Find(&history).Error
	return history, err
}

func (r *pointRepo) CreatePointTotal(pt *entity.PointTotal) error {
	return r.db.Create(pt).Error
}

func (r *pointRepo) GetAllTotalsWithUser() ([]entity.PointTotal, error) {
	var totals []entity.PointTotal
	err := r.db.Preload("User").Find(&totals).Error
	return totals, err
}
