package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	delivery "ybg-backend-go/internal/delivery/http" // Fixed Typo & Added Alias
	"ybg-backend-go/internal/delivery/http/middleware"
	"ybg-backend-go/internal/repository"
	"ybg-backend-go/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var router *gin.Engine

func init() {
	_ = godotenv.Load()

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Println("Warning: DB_URL is not set")
		return
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	repository.SeedAdmin(db)

	productRepo := repository.NewProductRepository(db)
	productUC := usecase.NewProductUsecase(productRepo)
	productHandler := delivery.NewProductHandler(productUC) // Use Alias

	userRepo := repository.NewUserRepository(db)
	pointRepo := repository.NewPointRepository(db)
	userUC := usecase.NewUserUsecase(userRepo, pointRepo)
	userHandler := delivery.NewUserHandler(userUC) // Use Alias

	newsRepo := repository.NewNewsRepository(db)
	newsUC := usecase.NewNewsUsecase(newsRepo)
	newsHandler := delivery.NewNewsHandler(newsUC) // Use Alias

	brandRepo := repository.NewBrandRepository(db)
	brandUC := usecase.NewBrandUsecase(brandRepo)
	brandHandler := delivery.NewBrandHandler(brandUC) // Use Alias

	categoryRepo := repository.NewCategoryRepository(db)
	categoryUC := usecase.NewCategoryUsecase(categoryRepo)
	categoryHandler := delivery.NewCategoryHandler(categoryUC) // Use Alias

	pRepo := repository.NewPointRepository(db)
	pUC := usecase.NewPointUsecase(pRepo)
	pHandler := delivery.NewPointHandler(pUC) // Use Alias

	r := gin.Default()

	r.POST("/register", userHandler.Create)
	r.POST("/login", userHandler.Login)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "database": "connected"})
	})

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		brandAdmin := api.Group("/brand")
		brandAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			brandAdmin.POST("/", brandHandler.Create)
			brandAdmin.DELETE("/:id", brandHandler.Delete)
		}

		categoryAdmin := api.Group("/category")
		categoryAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			categoryAdmin.POST("/", categoryHandler.Create)
			categoryAdmin.DELETE("/:id", categoryHandler.Delete)
		}

		api.GET("/products", productHandler.GetAll)
		api.GET("/products/:id", productHandler.GetByID)
		productAdmin := api.Group("/products")
		productAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			productAdmin.POST("/", productHandler.Create)
			productAdmin.PUT("/:id", productHandler.Update)
			productAdmin.DELETE("/:id", productHandler.Delete)
		}

		points := api.Group("/points")
		{
			points.GET("/history", pHandler.GetHistory)
			points.POST("/", middleware.RoleMiddleware("admin"), pHandler.CreatePoint)
			points.GET("/all", middleware.RoleMiddleware("admin"), pHandler.GetAllSummaries)
		}

		api.GET("/news", newsHandler.GetAll)
		newsAdmin := api.Group("/news")
		newsAdmin.Use(middleware.RoleMiddleware("admin"))
		{
			newsAdmin.POST("/", newsHandler.Create)
			newsAdmin.PUT("/:id", newsHandler.Update)
			newsAdmin.DELETE("/:id", newsHandler.Delete)
		}

		api.GET("/users", middleware.RoleMiddleware("admin"), userHandler.GetAll)
		api.GET("/profile", userHandler.GetByID)
		api.PUT("/profile", userHandler.Update)
	}

	router = r
}

func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running locally on port %s...\n", port)
	router.Run(":" + port)
}
