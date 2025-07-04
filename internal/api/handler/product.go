package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// GetProducts - Tüm ürünleri getir
func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get products endpoint - henüz implement edilmedi",
	})
}

// GetProduct - Tekil ürün getir
func GetProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Get product endpoint - henüz implement edilmedi",
		"id":      id,
	})
}

// CreateProduct - Yeni ürün oluştur (Admin)
func CreateProduct(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Create product endpoint - henüz implement edilmedi",
	})
}

// UpdateProduct - Ürün güncelle (Admin)
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update product endpoint - henüz implement edilmedi",
		"id":      id,
	})
}

// DeleteProduct - Ürün sil (Admin)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete product endpoint - henüz implement edilmedi",
		"id":      id,
	})
}
