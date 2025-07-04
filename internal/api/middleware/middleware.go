package middleware

import (
	"fmt"
	"net/http"
	"strings"
	
	"ecommerce/internal/auth"
	"github.com/gin-gonic/gin"
)

// CORS middleware
func CORS() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// Logger middleware
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// AuthRequired middleware - JWT doğrulama için
func AuthRequired(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Authorization header'ından token'ı al
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header gerekli",
			})
			c.Abort()
			return
		}

		// "Bearer " prefix'ini kaldır
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Bearer token formatı gerekli",
			})
			c.Abort()
			return
		}

		// Token'ı doğrula
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Geçersiz token",
			})
			c.Abort()
			return
		}

		// Kullanıcı bilgilerini context'e ekle
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	})
}

// AdminRequired middleware - Admin yetkisi kontrolü
func AdminRequired() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Kullanıcı rolünü kontrol et
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Yetkilendirme bilgisi bulunamadı",
			})
			c.Abort()
			return
		}

		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Bu işlem için admin yetkisi gerekli",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}
