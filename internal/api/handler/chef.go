package handler

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

// GetChefs - Tüm şefleri getir
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
