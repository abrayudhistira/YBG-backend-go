package repository

import (
	"ybg-backend-go/core/entity"

	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(p *entity.Product) error
	GetAll() ([]entity.Product, error)
	GetByID(id uint) (entity.Product, error)
	Update(p *entity.Product) error
	Delete(id uint) error
}

type productRepo struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepo{db: db}
}

func (r *productRepo) Create(p *entity.Product) error {
	return r.db.Create(p).Error
}

func (r *productRepo) GetAll() ([]entity.Product, error) {
	var products []entity.Product
	// Mengambil data produk beserta info Brand dan Category
	err := r.db.Preload("Brand").Preload("Category").Find(&products).Error
	return products, err
}

func (r *productRepo) GetByID(id uint) (entity.Product, error) {
	var product entity.Product
	err := r.db.Preload("Brand").Preload("Category").First(&product, id).Error
	return product, err
}

//	func (r *productRepo) Update(p *entity.Product) error {
//		// Save akan memperbarui semua kolom berdasarkan Primary Key
//		return r.db.Save(p).Error
//	}
func (r *productRepo) Update(p *entity.Product) error {
	// Menggunakan Updates agar hanya field yang dikirim saja yang diperbarui
	return r.db.Model(p).Where("product_id = ?", p.ProductID).Updates(p).Error
}
func (r *productRepo) Delete(id uint) error {
	return r.db.Delete(&entity.Product{}, id).Error
}
