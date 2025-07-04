package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// GetProfile - Kullanıcı profil bilgilerini getir
func GetProfile(c *gin.Context) {
	// JWT middleware'den kullanıcı ID'sini al
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	// Profil bilgilerini getir
	user, err := deps.UserService.GetProfile(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Profil bilgileri alınamadı",
		})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Kullanıcı bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil bilgileri başarıyla getirildi",
		"data": user,
	})
}

// UpdateProfile - Kullanıcı profil bilgilerini güncelle
func UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Update profile endpoint - henüz implement edilmedi",
	})
}
