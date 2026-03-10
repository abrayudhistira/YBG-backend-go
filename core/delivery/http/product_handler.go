package http

import (
	"errors"
	"net/http"
	"strconv"
	"ybg-backend-go/core/entity"
	"ybg-backend-go/core/usecase"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct {
	uc usecase.ProductUsecase
}

func NewProductHandler(uc usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc: uc}
}

func (h *ProductHandler) RegisterRoutes(r *gin.Engine) {
	routes := r.Group("/products")
	{
		routes.POST("/", h.Create)
		routes.GET("/", h.GetAll)
		routes.GET("/:id", h.GetByID)
		routes.PUT("/:id", h.Update)
		routes.DELETE("/:id", h.Delete)
	}
}

// 1. GET ALL: Tambahkan validasi jika array kosong
func (h *ProductHandler) GetAll(c *gin.Context) {
	products, err := h.uc.FetchProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products", "details": err.Error()})
		return
	}

	// Jika data kosong, return array kosong [] bukan null, atau beri pesan
	if len(products) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No products found",
			"data":    []entity.Product{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products retrieved successfully",
		"count":   len(products),
		"data":    products,
	})
}

// 2. GET BY ID: Cek apakah data benar-benar ada
func (h *ProductHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format. ID must be a positive integer"})
		return
	}

	product, err := h.uc.GetProductDetail(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found", "id": id})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product detail retrieved",
		"data":    product,
	})
}

// 3. CREATE: Validasi field wajib
func (h *ProductHandler) Create(c *gin.Context) {
	var p entity.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validasi manual sederhana (bisa ditingkatkan dengan library validator)
	if p.Name == "" || p.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and Price are required and must be valid"})
		return
	}

	if err := h.uc.CreateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"data":    p,
	})
}

// 4. UPDATE: Cek keberadaan data sebelum update
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var p entity.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p.ProductID = uint(id)
	if err := h.uc.UpdateProduct(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"data":    p,
	})
}

// 5. DELETE
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.uc.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product with ID " + strconv.Itoa(id) + " has been deleted",
	})
}
