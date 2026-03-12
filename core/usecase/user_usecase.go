package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	SyncWithSpreadsheet() (map[string]interface{}, error)
	SyncUpsert(user *entity.User) (string, error)
	SyncClean(activeIDs []string) (int64, error)
	StoreTemporaryOTP(email, otp string) error
	ValidateOTPAndGenerateResetToken(email, otp string) (string, error)
	ResetPasswordWithToken(token, newPassword string) error
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

func (u *userUC) SyncWithSpreadsheet() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	spreadsheetID := os.Getenv("SPREADSHEET_ID")
	rangeName := os.Getenv("SHEET_RANGE")

	srv, err := utils.GetSheetsService(ctx)
	if err != nil {
		return nil, fmt.Errorf("Google Service Error: %v", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return nil, fmt.Errorf("Sheet Data Error: %v", err)
	}

	var successCount, failCount int
	var errorDetails []string

	for i, row := range resp.Values {
		if i == 0 {
			continue
		} // Skip header jika ada

		if len(row) < 5 { // Minimal kolom yang dibutuhkan
			failCount++
			errorDetails = append(errorDetails, fmt.Sprintf("Baris %d: Kolom kurang", i+1))
			continue
		}

		userID, _ := row[0].(string)
		fullName, _ := row[1].(string) // Sesuaikan indeks kolommu

		user := &entity.User{
			UserID:   userID,
			Name:     fullName,
			Email:    strings.ToLower(userID) + "@ybg.com",
			Password: "password123",
			Role:     "customer",
		}

		// Eksekusi Register
		if err := u.RegisterUser(user); err != nil {
			failCount++
			errMsg := fmt.Sprintf("Baris %d (ID: %s): %v", i+1, userID, err)
			errorDetails = append(errorDetails, errMsg)
			fmt.Println("[Sync Error]", errMsg) // Muncul di Log Vercel
			continue
		}
		successCount++
	}

	return map[string]interface{}{
		"total_data": len(resp.Values) - 1,
		"success":    successCount,
		"failed":     failCount,
		"errors":     errorDetails,
	}, nil
}
func (u *userUC) SyncUpsert(user *entity.User) (string, error) {
	// 1. Cek apakah UserID sudah ada di Repo
	existing, err := u.repo.GetByID(user.UserID)

	if err != nil {
		// JIKA TIDAK ADA -> Register Baru
		// Menggunakan RegisterUser yang sudah kamu punya (otomatis hashing & point)
		errReg := u.RegisterUser(user)
		if errReg != nil {
			return "", errReg
		}
		return "created", nil
	}

	// JIKA ADA -> Update data yang berubah
	existing.Name = user.Name
	existing.Email = user.Email
	existing.Phone = user.Phone
	existing.Birth = user.Birth
	// Tambahkan field lain jika perlu diupdate

	errUpdate := u.repo.Update(&existing)
	if errUpdate != nil {
		return "", errUpdate
	}
	return "updated", nil
}

func (u *userUC) SyncClean(activeIDs []string) (int64, error) {
	return u.repo.DeleteNotIn(activeIDs, "customer")
}

func (u *userUC) StoreTemporaryOTP(email, otp string) error {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return errors.New("email tidak terdaftar")
	}

	user.OTPCode = otp
	// u.repo.Update sudah kita setting Select-nya untuk mengizinkan OTPCode
	return u.repo.Update(&user)
}

func (u *userUC) ValidateOTPAndGenerateResetToken(email, otp string) (string, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	if user.OTPCode == "" || user.OTPCode != otp {
		return "", errors.New("OTP salah atau sudah kadaluarsa")
	}

	resetToken := fmt.Sprintf("RST-%s-%s", user.UserID, utils.GenerateRandomID(10))
	expiry := time.Now().Add(15 * time.Minute)

	user.ResetToken = resetToken
	user.TokenExpiredAt = &expiry
	user.OTPCode = "" // Burn after use

	errUpdate := u.repo.Update(&user)
	if errUpdate != nil {
		return "", errUpdate
	}

	return resetToken, nil
}

func (u *userUC) ResetPasswordWithToken(token, newPassword string) error {
	// 1. Cari user berdasarkan ResetToken via Repo
	user, err := u.repo.GetByResetToken(token)
	if err != nil {
		return errors.New("token tidak valid atau tidak ditemukan")
	}

	// 2. Cek Expiry (menggunakan pointer check)
	if user.TokenExpiredAt == nil || time.Now().After(*user.TokenExpiredAt) {
		return errors.New("sesi reset password sudah habis, silakan ulangi dari Telegram")
	}

	// 3. Hash Password Baru
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("gagal memproses password baru")
	}

	// 4. Update via Repo (Menggunakan fungsi UpdatePassword yang sudah kamu buat di repo)
	// Fungsi ini otomatis membersihkan Token, OTP, dan ExpiredAt
	err = u.repo.UpdatePassword(user.UserID, hashed)
	if err != nil {
		return errors.New("gagal mengupdate password di database")
	}

	return nil
}
