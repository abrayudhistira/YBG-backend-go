package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
	SyncWithSpreadsheet() error
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
	if user.UserID != "" {
		existingID, _ := u.repo.GetByID(user.UserID)
		if existingID.UserID != "" {
			return errors.New("conflict: id already exists")
		}
	} else {
		user.UserID = utils.GenerateRandomID(8)
	}

	// 2. Cek duplikasi Email
	existingEmail, _ := u.repo.GetByEmail(user.Email)
	if existingEmail.Email != "" {
		return errors.New("conflict: email already exists")
	}

	// 3. Cek duplikasi Nama
	existingName, _ := u.repo.GetByName(user.Name) // Pastikan fungsi ini ada di Repo
	if existingName.Name != "" {
		return errors.New("conflict: name already exists")
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

func (u *userUC) SyncWithSpreadsheet() error {
	ctx := context.Background()
	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	rangeName := os.Getenv("SHEET_RANGE")
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if spreadsheetID == "" || credsPath == "" {
		return errors.New("konfigurasi spreadsheet di .env belum lengkap")
	}

	srv, err := utils.GetSheetsService(ctx, credsPath)
	if err != nil {
		return err
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return err
	}

	if len(resp.Values) == 0 {
		return errors.New("tidak ada data ditemukan di spreadsheet")
	}

	for i, row := range resp.Values {
		// 1. Validasi Panjang Kolom (Pastikan sampai indeks ke-8 ada)
		if len(row) < 9 {
			log.Printf("Baris %d di-skip: jumlah kolom kurang", i+2) // i+2 karena index 0 & header
			continue
		}

		// 2. Safe Type Assertion (Mencegah panic jika sel bukan string atau nil)
		userID, _ := row[0].(string)
		fullName, _ := row[8].(string)
		phone, _ := row[4].(string)
		birthDateStr, _ := row[3].(string)

		// 3. Validasi Data Wajib
		if userID == "" || fullName == "" {
			log.Printf("Baris %d di-skip: ID atau Nama kosong", i+2)
			continue
		}

		// 4. Parsing Tanggal dengan Check Error
		var birthDatePtr *time.Time
		t, errDate := time.Parse("2006-01-02", birthDateStr)
		if errDate == nil {
			birthDatePtr = &t
		}

		user := &entity.User{
			UserID:   userID,
			Name:     fullName,
			Birth:    birthDatePtr,
			Phone:    phone,
			Email:    strings.ToLower(userID) + "@ybg.com",
			Password: "password123",
			Role:     "customer",
		}

		// 5. Eksekusi dengan Triple Check yang sudah ada
		if err := u.RegisterUser(user); err != nil {
			log.Printf("Skip user %s: %v", userID, err)
			continue
		}
	}

	return nil
}
