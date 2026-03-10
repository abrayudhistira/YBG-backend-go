package usecase

import (
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type ProductUsecase interface {
	CreateProduct(p *entity.Product) error
	FetchProducts() ([]entity.Product, error)
	GetProductDetail(id uint) (entity.Product, error)
	UpdateProduct(p *entity.Product) error
	DeleteProduct(id uint) error
}

type productUC struct {
	repo repository.ProductRepository
}

func NewProductUsecase(repo repository.ProductRepository) ProductUsecase {
	return &productUC{repo: repo}
}

func (u *productUC) CreateProduct(p *entity.Product) error {
	return u.repo.Create(p)
}

func (u *productUC) FetchProducts() ([]entity.Product, error) {
	return u.repo.GetAll()
}

func (u *productUC) GetProductDetail(id uint) (entity.Product, error) {
	return u.repo.GetByID(id)
}

func (u *productUC) UpdateProduct(p *entity.Product) error {
	return u.repo.Update(p)
}

func (u *productUC) DeleteProduct(id uint) error {
	return u.repo.Delete(id)
}
