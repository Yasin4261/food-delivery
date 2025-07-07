package handler

import (
	"net/http"
	"strconv"
	
	"github.com/gin-gonic/gin"
)

// GetMeals - Tüm yemekleri getir
func GetMeals(c *gin.Context) {
	deps := GetDependencies()
	meals, err := deps.MealService.GetAllMeals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Yemekler getirilemedi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meals": meals,
	})
}

// GetMeal - Tekil yemek getir
func GetMeal(c *gin.Context) {
	id := c.Param("id")
	mealID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Geçersiz yemek ID",
		})
		return
	}

	deps := GetDependencies()
	meal, err := deps.MealService.GetMealByID(uint(mealID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Yemek bulunamadı",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meal": meal,
	})
}

// GetMyMeals - Şefin kendi yemeklerini getir
func GetMyMeals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Kullanıcı bilgisi bulunamadı",
		})
		return
	}

	deps := GetDependencies()
	meals, err := deps.MealService.GetMealsByChefID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Yemekler getirilemedi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meals": meals,
	})
}

// CreateMeal - Yeni yemek oluştur (Chef)
func CreateMeal(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Create meal endpoint - henüz implement edilmedi",
	})
}

// UpdateMeal - Yemek güncelle (Chef)
func UpdateMeal(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Update meal endpoint - henüz implement edilmedi",
		"id":      id,
	})
}

// DeleteMeal - Yemek sil (Chef)
func DeleteMeal(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Delete meal endpoint - henüz implement edilmedi",
		"id":      id,
	})
}

// ToggleMealAvailability - Yemek durumunu değiştir (Chef)
func ToggleMealAvailability(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"message": "Toggle meal availability endpoint - henüz implement edilmedi",
		"id":      id,
	})
}
