package wire

import (
	"ybg-backend-go/core/delivery/http"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/core/usecase"

	"gorm.io/gorm"
)

// InitializeUserUsecaseManual: Rakit usecase secara manual tanpa generator
func InitializeUserUsecaseManual(db *gorm.DB) usecase.UserUsecase {
	userRepo := repository.NewUserRepository(db)
	pointRepo := repository.NewPointRepository(db)
	return usecase.NewUserUsecase(userRepo, pointRepo)
}

// NewUserHandlerManual: Masukkan usecase ke dalam handler
func NewUserHandlerManual(uc usecase.UserUsecase) *http.UserHandler {
	return http.NewUserHandler(uc)
}