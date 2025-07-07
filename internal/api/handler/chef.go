package handler

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

// GetChefs - Tüm şefleri getir
// @Summary Tüm şefleri listele
// @Description Platformdaki tüm aktif şefleri getirir
// @Tags Chefs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Şefler başarıyla getirildi"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /chefs [get]
func GetChefs(c *gin.Context) {
	deps := GetDependencies()
	chefs, err := deps.ChefService.GetAllChefs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Şefler getirilemedi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chefs": chefs,
	})
}

// GetChef - Tekil şef getir
// @Summary Belirli bir şefi getir
// @Description ID'ye göre şef detaylarını getirir
// @Tags Chefs
// @Accept json
// @Produce json
// @Param id path int true "Şef ID"
// @Success 200 {object} map[string]interface{} "Şef başarıyla getirildi"
// @Failure 400 {object} map[string]string "Geçersiz ID"
// @Failure 404 {object} map[string]string "Şef bulunamadı"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /chefs/{id} [get]
func GetChef(c *gin.Context) {
	id := c.Param("id")
	chefID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz şef ID",
		})
		return
	}

	deps := GetDependencies()
	chef, err := deps.ChefService.GetChefByID(uint(chefID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Şef bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chef": chef,
	})
}

// GetChefMeals - Şefin yemeklerini getir
// @Summary Şefin yemeklerini listele
// @Description Belirli bir şefin tüm yemeklerini getirir
// @Tags Chefs
// @Accept json
// @Produce json
// @Param id path int true "Şef ID"
// @Success 200 {object} map[string]interface{} "Şefin yemekleri başarıyla getirildi"
// @Failure 400 {object} map[string]string "Geçersiz ID"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /chefs/{id}/meals [get]
func GetChefMeals(c *gin.Context) {
	id := c.Param("id")
	chefID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz şef ID",
		})
		return
	}

	deps := GetDependencies()
	meals, err := deps.MealService.GetMealsByChefID(uint(chefID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Şefin yemekleri getirilemedi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meals": meals,
	})
}

// GetChefProfile - Şef profilini getir
// @Summary Şef profilini getir
// @Description Oturum açmış şefin profil bilgilerini getirir
// @Tags Chef Profile
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Şef profili başarıyla getirildi"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 404 {object} map[string]string "Şef profili bulunamadı"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /chef/profile [get]
func GetChefProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	deps := GetDependencies()
	chef, err := deps.ChefService.GetChefByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Şef profili bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chef": chef,
	})
}

// CreateChefProfile - Şef profili oluştur
func CreateChefProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Create chef profile endpoint - henüz implement edilmedi",
	})
}

// UpdateChefProfile - Şef profili güncelle
func UpdateChefProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Update chef profile endpoint - henüz implement edilmedi",
	})
}

// GetChefOrders - Şefin siparişlerini getir
func GetChefOrders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Get chef orders endpoint - henüz implement edilmedi",
	})
}
