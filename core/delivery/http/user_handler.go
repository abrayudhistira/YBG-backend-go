package http

import (
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"
	"ybg-backend-go/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	uc usecase.UserUsecase
}

func NewUserHandler(uc usecase.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	routes := r.Group("/users")
	{
		routes.POST("/", h.Create)
		routes.GET("/", h.GetAll)
		routes.GET("/:id", h.GetByID)
		routes.PUT("/:id", h.Update)
		routes.DELETE("/:id", h.Delete)
		routes.POST("/login", h.Login)
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var u entity.User
	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format input salah", "details": err.Error()})
		return
	}

	// 1. Sanitasi & Validasi Basic
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
	u.Name = strings.TrimSpace(u.Name)

	if u.Email == "" || u.Password == "" || u.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama, Email, dan Password wajib diisi"})
		return
	}

	// 2. Validasi Format Email
	if _, err := mail.ParseAddress(u.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	// 3. Validasi Panjang Password
	if len(u.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password minimal 6 karakter"})
		return
	}

	// 4. Eksekusi ke Usecase
	// if err := h.uc.RegisterUser(&u); err != nil {
	// 	// Cek jika error karena duplikasi email (asumsi usecase/repo mengembalikan error spesifik)
	// 	if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
	// 		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
	// 		return
	// 	}
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal registrasi user"})
	// 	return
	// }
	if err := h.uc.RegisterUser(&u); err != nil {
		errStr := strings.ToLower(err.Error())

		switch {
		case strings.Contains(errStr, "id already exists"):
			c.JSON(http.StatusConflict, gin.H{"error": "ID sudah terdaftar"})
		case strings.Contains(errStr, "email already exists"):
			c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		case strings.Contains(errStr, "name already exists"):
			c.JSON(http.StatusConflict, gin.H{"error": "Nama sudah terdaftar"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal registrasi user"})
		}
		return
	}

	// Jangan kirim balik password di response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User berhasil dibuat",
		"data": gin.H{
			"user_id": u.UserID,
			"name":    u.Name,
			"email":   u.Email,
		},
	})
}

func (h *UserHandler) GetAll(c *gin.Context) {
	users, err := h.uc.FetchAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No users found", "data": []entity.User{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	id := c.Param("id") // String ID (bisa UUID string atau Y001)

	user, err := h.uc.GetUserProfile(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// 1. Ambil data dari form
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(strings.ToLower(c.PostForm("email")))

	if name == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan Email tidak boleh kosong"})
		return
	}

	// 2. Validasi Format Email
	if _, err := mail.ParseAddress(email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	var u entity.User
	u.UserID = id
	u.Name = name
	u.Email = email

	// 3. Handling File Gambar
	var imageStream io.Reader
	var fileName, contentType string

	file, err := c.FormFile("image")
	if err == nil {
		if file.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran gambar maksimal 5MB"})
			return
		}

		// Validasi tipe file (MIME type)
		ext := strings.ToLower(file.Header.Get("Content-Type"))
		if !strings.HasPrefix(ext, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File harus berupa gambar (jpg/png)"})
			return
		}

		openedFile, _ := file.Open()
		defer openedFile.Close()
		imageStream = openedFile
		fileName = file.Filename
		contentType = file.Header.Get("Content-Type")
	}

	if err := h.uc.UpdateProfile(&u, imageStream, fileName, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal update profil"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui",
		"data": gin.H{
			"user_id":         u.UserID,
			"name":            u.Name,
			"email":           u.Email,
			"profile_picture": u.ProfilePicture,
		},
	})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.uc.RemoveUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
		return
	}

	user, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := utils.GenerateToken(user.UserID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user": gin.H{
			"user_id": user.UserID,
			"name":    user.Name,
			"email":   user.Email,
			"role":    user.Role,
		},
	})
}
