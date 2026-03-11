package usecase

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/repository"
	"ybg-backend-go/pkg/utils"
)

type UserUsecase interface {
	RegisterUser(u *entity.User) error
	FetchAllUsers() ([]entity.User, error)
	GetUserProfile(id string) (entity.User, error)
	UpdateProfile(u *entity.User, file io.Reader, fileName, contentType string) error
	RemoveUser(id string) error
	Login(email, password string) (entity.User, error)
}

type userUC struct {
	repo      repository.UserRepository
	pointRepo repository.PointRepository
}

func NewUserUsecase(repo repository.UserRepository, pointRepo repository.PointRepository) UserUsecase {
	return &userUC{repo: repo, pointRepo: pointRepo}
}

func (u *userUC) UpdateProfile(user *entity.User, file io.Reader, fileName, contentType string) error {
	if file != nil {
		supabaseURL := os.Getenv("SUPABASE_URL")
		supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		bucketName := "avatars"

		// PERBAIKAN: Karena UserID sudah string, hapus method .String()
		remotePath := fmt.Sprintf("%s/%s", user.UserID, fileName)
		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, remotePath)

		buf := new(bytes.Buffer)
		buf.ReadFrom(file)

		req, _ := http.NewRequest("POST", uploadURL, buf)
		req.Header.Set("Authorization", "Bearer "+supabaseKey)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("x-upsert", "true")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == 201) {
			user.ProfilePicture = fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseURL, bucketName, remotePath)
		}
		if resp != nil {
			defer resp.Body.Close()
		}
	}

	return u.repo.Update(user)
}

func (u *userUC) RegisterUser(user *entity.User) error {
	existingUser, _ := u.repo.GetByEmail(user.Email)
	if existingUser.Email != "" {
		return errors.New("duplicate: email already exists")
	}
	// PERBAIKAN: Pastikan fungsi di utils sudah sesuai namanya
	if user.UserID == "" {
		user.UserID = utils.GenerateRandomID(8) // Pastikan nama ini sama dengan di pkg/utils
	}

	hashed, _ := utils.HashPassword(user.Password)
	user.Password = hashed

	if err := u.repo.Create(user); err != nil {
		return err
	}

	return u.pointRepo.CreatePointTotal(&entity.PointTotal{
		UserID: user.UserID, Total: 0, Tier: "friend", CreatedAt: time.Now(),
	})
}

func (u *userUC) FetchAllUsers() ([]entity.User, error) {
	return u.repo.GetAll()
}

// PERBAIKAN: Error "cannot use id as uuid.UUID" berarti REPOSITORY Anda masih pakai UUID.
// Anda HARUS mengubah interface UserRepository.GetByID agar menerima string.
func (u *userUC) GetUserProfile(id string) (entity.User, error) {
	return u.repo.GetByID(id)
}

func (u *userUC) RemoveUser(id string) error {
	return u.repo.Delete(id)
}

func (u *userUC) Login(email, password string) (entity.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil || !utils.CheckPasswordHash(password, user.Password) {
		return entity.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
