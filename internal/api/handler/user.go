package handler

import (
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// GetProfile - Kullanıcı profil bilgilerini getir
// @Summary Kullanıcı profilini getir
// @Description Oturum açmış kullanıcının profil bilgilerini getirir
// @Tags User Profile
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Profil başarıyla getirildi"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 404 {object} map[string]string "Kullanıcı bulunamadı"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /profile [get]
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
// @Summary Kullanıcı profilini güncelle
// @Description Oturum açmış kullanıcının profil bilgilerini günceller
// @Tags User Profile
// @Accept json
// @Produce json
// @Security Bearer
// @Param profile body map[string]interface{} true "Güncellenecek profil bilgileri"
// @Success 200 {object} map[string]interface{} "Profil başarıyla güncellendi"
// @Failure 400 {object} map[string]string "Geçersiz istek"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /profile [put]
func UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Update profile endpoint - henüz implement edilmedi",
	})
}
