package handler

import (
	"net/http"
	
	"ecommerce/internal/model"
	"github.com/gin-gonic/gin"
)

// Login - Kullanıcı giriş handler'ı
func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
			"details": err.Error(),
		})
		return
	}

	// Login işlemi
	authResponse, err := deps.UserService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Başarıyla giriş yapıldı",
		"data": authResponse,
	})
}

// Register - Kullanıcı kayıt handler'ı
func Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz istek formatı",
			"details": err.Error(),
		})
		return
	}

	// Register işlemi
	authResponse, err := deps.UserService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Hesap başarıyla oluşturuldu",
		"data": authResponse,
	})
}

// Logout - Kullanıcı çıkış handler'ı
func Logout(c *gin.Context) {
	// JWT tabanlı logout - client-side token silme
	// Sunucu tarafında herhangi bir işlem yapmaya gerek yok
	// Token blacklist implementasyonu isterseniz burada yapılabilir
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Başarıyla çıkış yapıldı",
	})
}
