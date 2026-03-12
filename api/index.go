package handler

import (
	"net/http"
	"os"

	"ybg-backend-go/core/delivery/http/middleware"
	"ybg-backend-go/core/wire"
	"ybg-backend-go/pkg/telegram" // Pastikan folder telegram kamu di-import

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var router *gin.Engine

func init() {
	// 1. Load ENV (Lokal)
	_ = godotenv.Load()

	// 2. Koneksi Database
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return
	}

	// 3. Ambil Token Bot dari Environment
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	// 4. Inisialisasi Usecase & Handler via Wire
	// Kita inisialisasi userUC secara mandiri agar bisa dipakai Bot & Handler
	userUC := wire.InitializeUserUsecaseManual(db)
	userHandler := wire.NewUserHandlerManual(userUC)

	// Inisialisasi Handler Lainnya
	productHandler := wire.InitializeProductHandler(db)
	newsHandler := wire.InitializeNewsHandler(db)
	brandHandler := wire.InitializeBrandHandler(db)
	categoryHandler := wire.InitializeCategoryHandler(db)
	pHandler := wire.InitializePointHandler(db)

	// 5. Inisialisasi Bot Service (Webhook Mode)
	botSvc := telegram.NewBotService(token, userUC)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// --- PUBLIC ROUTES ---
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/health")
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	// Route Auth Publik
	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login)
	r.POST("/reset-password", userHandler.ResetPassword)

	// --- TELEGRAM WEBHOOK ROUTE ---
	r.POST("/telegram-webhook", func(c *gin.Context) {
		var update tgbotapi.Update
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}
		// Bot memproses pesan yang masuk
		botSvc.HandleUpdate(update)
		c.Status(http.StatusOK)
	})

	// --- SYNC ROUTES (Biasanya butuh secret key) ---
	r.POST("/users/sync", userHandler.SyncSheets)
	r.POST("/users/sync-push", userHandler.SyncPush)
	r.POST("/users/sync-clean", userHandler.SyncClean)

	// --- PRIVATE API ROUTES (Auth Middleware) ---
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Profile
		api.GET("/profile/:id", userHandler.GetByID)
		api.PUT("/profile/:id", userHandler.Update)
		api.GET("/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)

		// Brand & Category
		brandAdmin := api.Group("/brand")
		api.GET("/brand", brandHandler.GetAll)
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/", brandHandler.Create)
			brandAdmin.DELETE("/:id", brandHandler.Delete)
		}

		categoryAdmin := api.Group("/category")
		api.GET("/category", categoryHandler.GetAll)
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/", categoryHandler.Create)
			categoryAdmin.DELETE("/:id", categoryHandler.Delete)
		}

		// Products
		api.GET("/products", productHandler.GetAll)
		api.GET("/products/:id", productHandler.GetByID)
		api.GET("/products/search", productHandler.Search)
		productAdmin := api.Group("/products")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/", productHandler.Create)
			productAdmin.PUT("/:id", productHandler.Update)
			productAdmin.DELETE("/:id", productHandler.Delete)
		}

		// Points
		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		// News
		api.GET("/news", newsHandler.GetAll)
		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/", newsHandler.Create)
			newsAdmin.PUT("/:id", newsHandler.Update)
			newsAdmin.DELETE("/:id", newsHandler.Delete)
		}
	}

	router = r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if router == nil {
		http.Error(w, "Router not initialized", http.StatusInternalServerError)
		return
	}
	router.ServeHTTP(w, r)
}
