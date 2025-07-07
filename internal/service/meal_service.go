package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// MealService - yemek iş mantığı
type MealService struct {
	mealRepo *repository.MealRepository
	chefRepo *repository.ChefRepository
}

func NewMealService(mealRepo *repository.MealRepository, chefRepo *repository.ChefRepository) *MealService {
	return &MealService{
		mealRepo: mealRepo,
		chefRepo: chefRepo,
	}
}

func (s *MealService) CreateMeal(userID uint, req *model.MealRequest) (*model.Meal, error) {
	// Chef profilini kontrol et
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if chef == nil {
		return nil, errors.New("chef profili bulunamadı")
	}
	if !chef.IsActive {
		return nil, errors.New("chef profili aktif değil")
	}
	
	// Yemek oluştur
	meal := &model.Meal{
		ChefID:            chef.ID,
		Name:              req.Name,
		Description:       req.Description,
		Category:          req.Category,
		Cuisine:           req.Cuisine,
		Price:             req.Price,
		Currency:          "TRY",
		Portion:           req.Portion,
		ServingSize:       req.ServingSize,
		AvailableQuantity: req.AvailableQuantity,
		PreparationTime:   req.PreparationTime,
		CookingTime:       req.CookingTime,
		Calories:          req.Calories,
		Ingredients:       req.Ingredients,
		Allergens:         req.Allergens,
		IsVegetarian:      req.IsVegetarian,
		IsVegan:           req.IsVegan,
		IsGlutenFree:      req.IsGlutenFree,
		IsActive:          true,
		IsAvailable:       true,
		Rating:            0,
		TotalOrders:       0,
	}
	
	err = s.mealRepo.Create(meal)
	if err != nil {
		return nil, err
	}
	
	return meal, nil
}

func (s *MealService) GetMealsByChef(chefID uint) ([]model.Meal, error) {
	return s.mealRepo.GetByChefID(chefID)
}

func (s *MealService) GetMealsByChefUser(userID uint) ([]model.Meal, error) {
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if chef == nil {
		return nil, errors.New("chef profili bulunamadı")
	}
	
	return s.mealRepo.GetByChefID(chef.ID)
}

func (s *MealService) GetMeal(mealID uint) (*model.Meal, error) {
	return s.mealRepo.GetByID(mealID)
}

func (s *MealService) UpdateMeal(userID uint, mealID uint, req *model.MealRequest) (*model.Meal, error) {
	// Chef kontrolü
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if chef == nil {
		return nil, errors.New("chef profili bulunamadı")
	}
	
	// Yemek kontrolü
	meal, err := s.mealRepo.GetByID(mealID)
	if err != nil {
		return nil, err
	}
	if meal == nil {
		return nil, errors.New("yemek bulunamadı")
	}
	if meal.ChefID != chef.ID {
		return nil, errors.New("bu yemeği güncelleme yetkiniz yok")
	}
	
	// Güncelle
	meal.Name = req.Name
	meal.Description = req.Description
	meal.Category = req.Category
	meal.Cuisine = req.Cuisine
	meal.Price = req.Price
	meal.Portion = req.Portion
	meal.ServingSize = req.ServingSize
	meal.AvailableQuantity = req.AvailableQuantity
	meal.PreparationTime = req.PreparationTime
	meal.CookingTime = req.CookingTime
	meal.Calories = req.Calories
	meal.Ingredients = req.Ingredients
	meal.Allergens = req.Allergens
	meal.IsVegetarian = req.IsVegetarian
	meal.IsVegan = req.IsVegan
	meal.IsGlutenFree = req.IsGlutenFree
	
	err = s.mealRepo.Update(meal)
	if err != nil {
		return nil, err
	}
	
	return meal, nil
}

func (s *MealService) DeleteMeal(userID uint, mealID uint) error {
	// Chef kontrolü
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if chef == nil {
		return errors.New("chef profili bulunamadı")
	}
	
	// Yemek kontrolü
	meal, err := s.mealRepo.GetByID(mealID)
	if err != nil {
		return err
	}
	if meal == nil {
		return errors.New("yemek bulunamadı")
	}
	if meal.ChefID != chef.ID {
		return errors.New("bu yemeği silme yetkiniz yok")
	}
	
	return s.mealRepo.Delete(mealID)
}

func (s *MealService) SearchMeals(filter *model.MealFilter) ([]model.MealWithChef, error) {
	return s.mealRepo.SearchWithChef(filter)
}

func (s *MealService) GetAvailableMeals() ([]model.MealWithChef, error) {
	return s.mealRepo.GetAvailableWithChef()
}

func (s *MealService) GetMealsByCategory(category string) ([]model.MealWithChef, error) {
	return s.mealRepo.GetByCategoryWithChef(category)
}

func (s *MealService) GetMealsByLocation(city, district string) ([]model.MealWithChef, error) {
	return s.mealRepo.GetByLocationWithChef(city, district)
}

func (s *MealService) ToggleMealAvailability(userID uint, mealID uint) error {
	// Chef kontrolü
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return err
	}
	if chef == nil {
		return errors.New("chef profili bulunamadı")
	}
	
	// Yemek kontrolü
	meal, err := s.mealRepo.GetByID(mealID)
	if err != nil {
		return err
	}
	if meal == nil {
		return errors.New("yemek bulunamadı")
	}
	if meal.ChefID != chef.ID {
		return errors.New("bu yemeği güncelleme yetkiniz yok")
	}
	
	meal.IsAvailable = !meal.IsAvailable
	return s.mealRepo.Update(meal)
}

func (s *MealService) GetAllMeals() ([]model.Meal, error) {
	return s.mealRepo.GetAll()
}

func (s *MealService) GetMealByID(mealID uint) (*model.Meal, error) {
	return s.mealRepo.GetByID(mealID)
}

func (s *MealService) GetMealsByChefID(chefID uint) ([]model.Meal, error) {
	return s.mealRepo.GetByChefID(chefID)
}
