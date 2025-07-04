package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// OrderService - sipariş iş mantığı
type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	cartRepo    *repository.CartRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, cartRepo *repository.CartRepository) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		cartRepo:    cartRepo,
	}
}

func (s *OrderService) CreateOrder(userID uint, req *model.CreateOrderRequest) (*model.Order, error) {
	// Validasyonlar
	if userID == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	if req.Address == "" {
		return nil, errors.New("teslimat adresi boş olamaz")
	}
	if len(req.Items) == 0 {
		return nil, errors.New("sipariş öğeleri boş olamaz")
	}
	
	// Stok kontrolü ve toplam fiyat hesaplama
	var total float64
	for _, item := range req.Items {
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, errors.New("ürün bulunamadı: " + string(rune(item.ProductID)))
		}
		if product.Stock < item.Quantity {
			return nil, errors.New("yeterli stok yok: " + product.Name)
		}
		total += product.Price * float64(item.Quantity)
	}
	
	// Sipariş oluştur
	order := &model.Order{
		UserID:  userID,
		Total:   total,
		Status:  "pending",
		Address: req.Address,
	}
	
	err := s.orderRepo.Create(order)
	if err != nil {
		return nil, err
	}
	
	// Sipariş öğelerini oluştur ve stok güncelle
	for _, item := range req.Items {
		product, _ := s.productRepo.GetByID(item.ProductID)
		
		orderItem := &model.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price, // Sipariş anındaki fiyat
		}
		
		err = s.orderRepo.CreateOrderItem(orderItem)
		if err != nil {
			return nil, err
		}
		
		// Stok güncelle
		product.Stock -= item.Quantity
		err = s.productRepo.Update(product)
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
		product, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			continue // Ürün bulunamazsa geç
		}
		
		product.Stock += item.Quantity
		s.productRepo.Update(product)
	}
	
	return s.UpdateOrderStatus(id, "cancelled")
}
