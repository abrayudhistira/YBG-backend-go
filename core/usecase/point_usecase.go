package usecase

import (
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
)

type PointUsecase interface {
	AddPointTransaction(uid string, point int) error
	GetMyPointHistory(uid string) ([]entity.PointHistory, error)
	FetchAllUsersPoints() ([]entity.PointTotal, error)
}

type pointUC struct {
	repo repository.PointRepository
}

func NewPointUsecase(repo repository.PointRepository) PointUsecase { return &pointUC{repo: repo} }

func (u *pointUC) AddPointTransaction(uid string, point int) error {
	history := entity.PointHistory{
		UserID: uid,
		Point:  point,
	}
	if err := u.repo.CreateHistory(&history); err != nil {
		return err
	}
	return u.repo.UpdateTotal(uid, point)
}

func (u *pointUC) GetMyPointHistory(uid string) ([]entity.PointHistory, error) {
	return u.repo.GetHistoryByUserID(uid)
}

func (u *pointUC) FetchAllUsersPoints() ([]entity.PointTotal, error) {
	return u.repo.GetAllTotalsWithUser()
}
