package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
	"time"
)

// AdminService - admin iş mantığı (Ev yemekleri platformu için)
type AdminService struct {
	userRepo  *repository.UserRepository
	chefRepo  *repository.ChefRepository
	mealRepo  *repository.MealRepository
	orderRepo *repository.OrderRepository
}

func NewAdminService(userRepo *repository.UserRepository, chefRepo *repository.ChefRepository, mealRepo *repository.MealRepository, orderRepo *repository.OrderRepository) *AdminService {
	return &AdminService{
		userRepo:  userRepo,
		chefRepo:  chefRepo,
		mealRepo:  mealRepo,
		orderRepo: orderRepo,
	}
}

// Dashboard statistics
func (s *AdminService) GetDashboardStats() (*model.DashboardStats, error) {
	// Kullanıcı sayıları
	allUsers, err := s.userRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	totalUsers := len(allUsers)
	totalChefs := 0
	totalCustomers := 0
	
	for _, user := range allUsers {
		switch user.Role {
		case "chef":
			totalChefs++
		case "customer":
			totalCustomers++
		}
	}
	
	// Chef sayıları
	allChefs, err := s.chefRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	activeChefs := 0
	for _, chef := range allChefs {
		if chef.IsActive {
			activeChefs++
		}
	}
	
	// Meal sayıları
	allMeals, err := s.mealRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	// Order sayıları ve revenue
	allOrders, err := s.orderRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	totalRevenue := 0.0
	pendingOrders := 0
	for _, order := range allOrders {
		if order.Status == "delivered" {
			totalRevenue += order.Total
		}
		if order.Status == "pending" {
			pendingOrders++
		}
	}
	
	stats := &model.DashboardStats{
		TotalUsers:     totalUsers,
		TotalChefs:     len(allChefs),
		TotalCustomers: totalCustomers,
		TotalOrders:    len(allOrders),
		TotalMeals:     len(allMeals),
		TotalRevenue:   totalRevenue,
		PendingOrders:  pendingOrders,
		ActiveChefs:    activeChefs,
		LastUpdated:    time.Now(),
	}
	
	return stats, nil
}

func (s *AdminService) GetAllUsers() ([]model.User, error) {
	return s.userRepo.GetAll()
}

func (s *AdminService) GetAllChefs() ([]model.Chef, error) {
	return s.chefRepo.GetAll()
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
	
	validRoles := []string{"customer", "chef", "admin"}
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

func (s *AdminService) VerifyChef(chefID uint) error {
	chef, err := s.chefRepo.GetByID(chefID)
	if err != nil {
		return err
	}
	if chef == nil {
		return errors.New("chef bulunamadı")
	}
	
	chef.IsVerified = true
	return s.chefRepo.Update(chef)
}

func (s *AdminService) DeactivateUser(id uint) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("kullanıcı bulunamadı")
	}
	
	user.IsActive = false
	return s.userRepo.Update(user)
}

func (s *AdminService) GetPendingChefs() ([]model.Chef, error) {
	allChefs, err := s.chefRepo.GetAll()
	if err != nil {
		return nil, err
	}
	
	var pendingChefs []model.Chef
	for _, chef := range allChefs {
		if !chef.IsVerified && chef.IsActive {
			pendingChefs = append(pendingChefs, chef)
		}
	}
	
	return pendingChefs, nil
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

func (s *AdminService) UpdateOrderStatus(orderID uint, status string) error {
	// Validate status
	validStatuses := []string{"pending", "confirmed", "shipped", "delivered", "cancelled"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	
	if !isValid {
		return errors.New("geçersiz sipariş durumu")
	}
	
	return s.orderRepo.UpdateStatus(orderID, status)
}

func (s *AdminService) GetAllMeals() ([]model.Meal, error) {
	return s.mealRepo.GetAll()
}

func (s *AdminService) ApproveMeal(mealID uint) error {
	meal, err := s.mealRepo.GetByID(mealID)
	if err != nil {
		return err
	}
	if meal == nil {
		return errors.New("yemek bulunamadı")
	}
	
	meal.IsAvailable = true
	return s.mealRepo.Update(meal)
}

func (s *AdminService) DeleteMeal(mealID uint) error {
	return s.mealRepo.Delete(mealID)
}
