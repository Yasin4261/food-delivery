package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// AdminService - admin iş mantığı
type AdminService struct {
	userRepo    *repository.UserRepository
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
}

func NewAdminService(userRepo *repository.UserRepository, orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *AdminService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.GetAll()
}

func (s *AdminService) GetUser(id uint) (*model.User, error) {
	if id == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}
	
	return user, nil
}

func (s *AdminService) UpdateUserRole(id uint, role string) error {
	if id == 0 {
		return errors.New("geçersiz kullanıcı ID")
	}
	
	validRoles := []string{"customer", "admin"}
	isValid := false
	for _, validRole := range validRoles {
		if role == validRole {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return errors.New("geçersiz rol")
	}
	
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("kullanıcı bulunamadı")
	}
	
	user.Role = role
	return s.userRepo.Update(user)
}

func (s *AdminService) GetAllOrders() ([]model.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *AdminService) GetOrdersByStatus(status string) ([]model.Order, error) {
	validStatuses := []string{"pending", "confirmed", "shipped", "delivered", "cancelled"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return nil, errors.New("geçersiz sipariş durumu")
	}
	
	return s.orderRepo.GetByStatus(status)
}

func (s *AdminService) GetDashboardStats() (*model.DashboardStats, error) {
	// Toplam kullanıcı sayısı
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	// Toplam sipariş sayısı
	orders, err := s.orderRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	// Toplam ürün sayısı
	products, err := s.productRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	// Toplam gelir hesaplama
	var totalRevenue float64
	for _, order := range orders {
		if order.Status == "delivered" {
			totalRevenue += order.Total
		}
	}
	
	// Bekleyen siparişler
	pendingOrders := 0
	for _, order := range orders {
		if order.Status == "pending" {
			pendingOrders++
		}
	}
	
	stats := &model.DashboardStats{
		TotalUsers:     len(users),
		TotalOrders:    len(orders),
		TotalProducts:  len(products),
		TotalRevenue:   totalRevenue,
		PendingOrders:  pendingOrders,
	}
	
	return stats, nil
}

func (s *AdminService) GetRecentOrders(limit int) ([]model.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	
	return s.orderRepo.GetRecent(limit)
}

func (s *AdminService) GetLowStockProducts(threshold int) ([]model.Product, error) {
	if threshold <= 0 {
		threshold = 10
	}
	
	return s.productRepo.GetLowStock(threshold)
}
