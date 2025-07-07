package handler

import (
	"net/http"
	"strconv"
	
	"ecommerce/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// AdminGetDashboard - Admin: Dashboard istatistikleri
// @Summary Admin dashboard istatistikleri
// @Description Admin için genel platform istatistiklerini getirir
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Dashboard istatistikleri başarıyla getirildi"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 403 {object} map[string]string "Admin yetkisi gerekli"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /admin/dashboard [get]
func (h *AdminHandler) AdminGetDashboard(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "İstatistikler alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// AdminGetUsers - Admin: Tüm kullanıcıları getir
// @Summary Admin: Tüm kullanıcıları listele
// @Description Admin için platform kullanıcılarını getirir
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Kullanıcılar başarıyla getirildi"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 403 {object} map[string]string "Admin yetkisi gerekli"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /admin/users [get]
func (h *AdminHandler) AdminGetUsers(c *gin.Context) {
	users, err := h.adminService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Kullanıcılar alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    users,
		"total":   len(users),
	})
}

// AdminGetChefs - Admin: Tüm chef'leri getir
// @Summary Admin: Tüm şefleri listele
// @Description Admin için platform şeflerini getirir
// @Tags Admin
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Şefler başarıyla getirildi"
// @Failure 401 {object} map[string]string "Yetkisiz erişim"
// @Failure 403 {object} map[string]string "Admin yetkisi gerekli"
// @Failure 500 {object} map[string]string "Sunucu hatası"
// @Router /admin/chefs [get]
func (h *AdminHandler) AdminGetChefs(c *gin.Context) {
	chefs, err := h.adminService.GetAllChefs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Chef'ler alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chefs,
		"total":   len(chefs),
	})
}

// AdminGetUser - Admin: Tekil kullanıcı getir
func (h *AdminHandler) AdminGetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz kullanıcı ID",
		})
		return
	}

	user, err := h.adminService.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Kullanıcı bulunamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
	})
}

// AdminVerifyChef - Admin: Chef'i onayla
func (h *AdminHandler) AdminVerifyChef(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz chef ID",
		})
		return
	}

	err = h.adminService.VerifyChef(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Chef onaylanamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chef başarıyla onaylandı",
	})
}

// AdminGetPendingChefs - Admin: Onay bekleyen chef'ler
func (h *AdminHandler) AdminGetPendingChefs(c *gin.Context) {
	chefs, err := h.adminService.GetPendingChefs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Bekleyen chef'ler alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    chefs,
		"total":   len(chefs),
	})
}

// AdminGetOrders - Admin: Tüm siparişleri getir
func (h *AdminHandler) AdminGetOrders(c *gin.Context) {
	orders, err := h.adminService.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Siparişler alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"total":   len(orders),
	})
}

// AdminUpdateOrderStatus - Admin: Sipariş durumunu güncelle
func (h *AdminHandler) AdminUpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz sipariş ID",
		})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz JSON formatı",
		})
		return
	}

	err = h.adminService.UpdateOrderStatus(uint(id), req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Sipariş durumu güncellenemedi",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sipariş durumu güncellendi",
	})
}

// AdminGetMeals - Admin: Tüm yemekleri getir
func (h *AdminHandler) AdminGetMeals(c *gin.Context) {
	meals, err := h.adminService.GetAllMeals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Yemekler alınamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    meals,
		"total":   len(meals),
	})
}

// AdminApproveMeal - Admin: Yemeği onayla
func (h *AdminHandler) AdminApproveMeal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz yemek ID",
		})
		return
	}

	err = h.adminService.ApproveMeal(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Yemek onaylanamadı",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Yemek başarıyla onaylandı",
	})
}

// AdminDeleteMeal - Admin: Yemek sil
func (h *AdminHandler) AdminDeleteMeal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Geçersiz yemek ID",
		})
		return
	}

	err = h.adminService.DeleteMeal(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Yemek silinemedi",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Yemek başarıyla silindi",
	})
}
