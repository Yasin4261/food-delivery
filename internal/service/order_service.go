package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// OrderService - sipariş iş mantığı
type OrderService struct {
	orderRepo *repository.OrderRepository
	mealRepo  *repository.MealRepository
	cartRepo  *repository.CartRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, mealRepo *repository.MealRepository, cartRepo *repository.CartRepository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		mealRepo:  mealRepo,
		cartRepo:  cartRepo,
	}
}

func (s *OrderService) CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	// Validasyonlar
	if userID == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	if req.DeliveryAddress == "" && req.DeliveryType == "delivery" {
		return nil, errors.New("teslimat için adres gerekli")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("sipariş öğeleri boş olamaz")
	}
	
	// Stok kontrolü ve toplam fiyat hesaplama
	var total float64
	for _, item := range req.Items {
		meal, err := s.mealRepo.GetByID(item.MealID)
		if err != nil {
			return nil, err
		}
		if meal == nil {
			return nil, errors.New("yemek bulunamadı: " + string(rune(item.MealID)))
		}
		if meal.AvailableQuantity < item.Quantity {
			return nil, errors.New("yeterli stok yok: " + meal.Name)
		}
		total += meal.Price * float64(item.Quantity)
	}
	
	// Sipariş oluştur - ana chef bilgisini ilk yemekten al
	var mainChefID uint
	if len(req.Items) > 0 {
		firstMeal, _ := s.mealRepo.GetByID(req.Items[0].MealID)
		if firstMeal != nil {
			mainChefID = firstMeal.ChefID
		}
	}
	
	order := &model.Order{
		UserID:          userID,
		ChefID:          mainChefID,
		Total:           total,
		Status:          "pending",
		DeliveryType:    req.DeliveryType,
		Address:         req.DeliveryAddress,
		DeliveryAddress: req.DeliveryAddress,
		PaymentMethod:   req.PaymentMethod,
		CustomerNote:    req.CustomerNote,
	}
	
	err := s.orderRepo.Create(order)
	if err != nil {
		return nil, err
	}
	
	// Sipariş öğelerini oluştur ve stok güncelle
	for _, item := range req.Items {
		meal, _ := s.mealRepo.GetByID(item.MealID)
		
		orderItem := &model.OrderItem{
			OrderID:             order.ID,
			MealID:              item.MealID,
			ChefID:              meal.ChefID,
			Quantity:            item.Quantity,
			Price:               meal.Price, // Sipariş anındaki fiyat
			SpecialInstructions: item.SpecialInstructions,
		}
		
		err = s.orderRepo.CreateOrderItem(orderItem)
		if err != nil {
			return nil, err
		}
		
		// Stok güncelle
		meal.AvailableQuantity -= item.Quantity
		err = s.mealRepo.Update(meal)
		if err != nil {
			return nil, err
		}
	}
	
	return order, nil
}

func (s *OrderService) GetUserOrders(userID uint) ([]model.Order, error) {
	if userID == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	
	return s.orderRepo.GetByUserID(userID)
}

func (s *OrderService) GetOrder(id uint) (*model.Order, error) {
	if id == 0 {
		return nil, errors.New("geçersiz sipariş ID")
	}
	
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("sipariş bulunamadı")
	}
	
	return order, nil
}

func (s *OrderService) GetOrderItems(orderID uint) ([]model.OrderItem, error) {
	if orderID == 0 {
		return nil, errors.New("geçersiz sipariş ID")
	}
	
	return s.orderRepo.GetOrderItems(orderID)
}

func (s *OrderService) UpdateOrderStatus(id uint, status string) error {
	if id == 0 {
		return errors.New("geçersiz sipariş ID")
	}
	
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
	
	return s.orderRepo.UpdateStatus(id, status)
}

func (s *OrderService) CancelOrder(id uint) error {
	order, err := s.GetOrder(id)
	if err != nil {
		return err
	}
	
	if order.Status != "pending" {
		return errors.New("sadece beklemede olan siparişler iptal edilebilir")
	}
	
	// Stokları geri yükle
	orderItems, err := s.GetOrderItems(id)
	if err != nil {
		return err
	}
	
	for _, item := range orderItems {
		meal, err := s.mealRepo.GetByID(item.MealID)
		if err != nil {
			continue // Yemek bulunamazsa geç
		}
		
		meal.AvailableQuantity += item.Quantity
		s.mealRepo.Update(meal)
	}
	
	return s.UpdateOrderStatus(id, "cancelled")
}
