package handler

import (
	"ecommerce/internal/service"
)

// HandlerDependencies - handler bağımlılıkları
type HandlerDependencies struct {
	UserService    *service.UserService
	ProductService *service.ProductService
	CartService    *service.CartService
	OrderService   *service.OrderService
	AdminService   *service.AdminService
}

var deps *HandlerDependencies

// SetDependencies - handler bağımlılıklarını ayarla
func SetDependencies(d *HandlerDependencies) {
	deps = d
}

// GetDependencies - handler bağımlılıklarını al
func GetDependencies() *HandlerDependencies {
	return deps
}
