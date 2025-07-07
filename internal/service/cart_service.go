package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// CartService - sepet iş mantığı
type CartService struct {
	cartRepo *repository.CartRepository
	mealRepo *repository.MealRepository
}

func NewCartService(cartRepo *repository.CartRepository, mealRepo *repository.MealRepository) *CartService {
	return &CartService{
		cartRepo: cartRepo,
		mealRepo: mealRepo,
	}
}

func (s *CartService) GetOrCreateCart(userID uint) (*model.Cart, error) {
	if userID == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	
	// Mevcut sepeti bul
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	
	// Sepet yoksa oluştur
	if cart == nil {
		cart = &model.Cart{
			UserID: userID,
		}
		err = s.cartRepo.Create(cart)
		if err != nil {
			return nil, err
		}
	}
	
	return cart, nil
}

func (s *CartService) AddItem(userID uint, req *model.AddToCartRequest) error {
	if userID == 0 {
		return errors.New("geçersiz kullanıcı ID")
	}
	if req.MealID == 0 {
		return errors.New("geçersiz yemek ID")
	}
	if req.Quantity <= 0 {
		return errors.New("miktar 0'dan büyük olmalıdır")
	}
	
	// Yemek kontrolü
	meal, err := s.mealRepo.GetByID(req.MealID)
	if err != nil {
		return err
	}
	if meal == nil {
		return errors.New("yemek bulunamadı")
	}
	if meal.AvailableQuantity < req.Quantity {
		return errors.New("yeterli stok yok")
	}
	
	// Sepeti al veya oluştur
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}
	
	// Mevcut öğeyi kontrol et
	existingItem, err := s.cartRepo.GetCartItem(cart.ID, req.MealID)
	if err != nil {
		return err
	}
	
	if existingItem != nil {
		// Mevcut öğeyi güncelle
		existingItem.Quantity += req.Quantity
		if existingItem.Quantity > meal.AvailableQuantity {
			return errors.New("yeterli stok yok")
		}
		return s.cartRepo.UpdateCartItem(existingItem)
	} else {
		// Yeni öğe ekle
		cartItem := &model.CartItem{
			CartID:              cart.ID,
			MealID:              req.MealID,
			ChefID:              meal.ChefID,
			Quantity:            req.Quantity,
			SpecialInstructions: req.SpecialInstructions,
		}
		return s.cartRepo.CreateCartItem(cartItem)
	}
}

func (s *CartService) UpdateItem(userID uint, mealID uint, quantity int) error {
	if userID == 0 {
		return errors.New("geçersiz kullanıcı ID")
	}
	if mealID == 0 {
		return errors.New("geçersiz yemek ID")
	}
	if quantity < 0 {
		return errors.New("miktar negatif olamaz")
	}
	
	// Sepeti al
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if cart == nil {
		return errors.New("sepet bulunamadı")
	}
	
	// Öğeyi bul
	item, err := s.cartRepo.GetCartItem(cart.ID, mealID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("sepet öğesi bulunamadı")
	}
	
	if quantity == 0 {
		// Öğeyi sil
		return s.cartRepo.DeleteCartItem(item.ID)
	}
	
	// Stok kontrolü
	meal, err := s.mealRepo.GetByID(mealID)
	if err != nil {
		return err
	}
	if meal.AvailableQuantity < quantity {
		return errors.New("yeterli stok yok")
	}
	
	// Öğeyi güncelle
	item.Quantity = quantity
	return s.cartRepo.UpdateCartItem(item)
}

func (s *CartService) RemoveItem(userID uint, mealID uint) error {
	if userID == 0 {
		return errors.New("geçersiz kullanıcı ID")
	}
	if mealID == 0 {
		return errors.New("geçersiz yemek ID")
	}
	
	// Sepeti al
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if cart == nil {
		return errors.New("sepet bulunamadı")
	}
	
	// Öğeyi bul ve sil
	item, err := s.cartRepo.GetCartItem(cart.ID, mealID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("sepet öğesi bulunamadı")
	}
	
	return s.cartRepo.DeleteCartItem(item.ID)
}

func (s *CartService) GetCartItems(userID uint) ([]model.CartItem, error) {
	if userID == 0 {
		return nil, errors.New("geçersiz kullanıcı ID")
	}
	
	// Sepeti al
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return []model.CartItem{}, nil // Boş sepet
	}
	
	return s.cartRepo.GetCartItems(cart.ID)
}

func (s *CartService) ClearCart(userID uint) error {
	if userID == 0 {
		return errors.New("geçersiz kullanıcı ID")
	}
	
	// Sepeti al
	cart, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if cart == nil {
		return nil // Zaten boş
	}
	
	return s.cartRepo.ClearCart(cart.ID)
}

func (s *CartService) GetCartTotal(userID uint) (float64, error) {
	items, err := s.GetCartItems(userID)
	if err != nil {
		return 0, err
	}
	
	var total float64
	for _, item := range items {
		// Get meal price from the related meal
		if item.Meal != nil {
			total += item.Meal.Price * float64(item.Quantity)
		}
	}
	
	return total, nil
}
