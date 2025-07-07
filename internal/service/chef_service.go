package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// ChefService - chef iş mantığı
type ChefService struct {
	chefRepo *repository.ChefRepository
	userRepo *repository.UserRepository
}

func NewChefService(chefRepo *repository.ChefRepository, userRepo *repository.UserRepository) *ChefService {
	return &ChefService{
		chefRepo: chefRepo,
		userRepo: userRepo,
	}
}

func (s *ChefService) CreateProfile(userID uint, req *model.ChefProfile) (*model.Chef, error) {
	// Kullanıcının chef olup olmadığını kontrol et
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}
	if user.Role != "chef" {
		return nil, errors.New("sadece chef rolündeki kullanıcılar chef profili oluşturabilir")
	}
	
	// Zaten chef profili var mı kontrol et
	existingChef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existingChef != nil {
		return nil, errors.New("chef profili zaten mevcut")
	}
	
	// Chef profili oluştur
	chef := &model.Chef{
		UserID:      userID,
		KitchenName: req.KitchenName,
		Description: req.Description,
		Speciality:  req.Speciality,
		Experience:  req.Experience,
		Address:     req.Address,
		District:    req.District,
		City:        req.City,
		IsActive:    true,
		IsVerified:  false,
		Rating:      0,
		TotalOrders: 0,
	}
	
	err = s.chefRepo.Create(chef)
	if err != nil {
		return nil, err
	}
	
	return chef, nil
}

func (s *ChefService) GetProfile(userID uint) (*model.Chef, error) {
	return s.chefRepo.GetByUserID(userID)
}

func (s *ChefService) UpdateProfile(userID uint, req *model.ChefProfile) (*model.Chef, error) {
	chef, err := s.chefRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if chef == nil {
		return nil, errors.New("chef profili bulunamadı")
	}
	
	// Güncelle
	chef.KitchenName = req.KitchenName
	chef.Description = req.Description
	chef.Speciality = req.Speciality
	chef.Experience = req.Experience
	chef.Address = req.Address
	chef.District = req.District
	chef.City = req.City
	
	err = s.chefRepo.Update(chef)
	if err != nil {
		return nil, err
	}
	
	return chef, nil
}

func (s *ChefService) GetChefsByLocation(city, district string) ([]model.Chef, error) {
	return s.chefRepo.GetByLocation(city, district)
}

func (s *ChefService) GetActiveChefs() ([]model.Chef, error) {
	return s.chefRepo.GetActiveChefs()
}

func (s *ChefService) GetChefWithMeals(chefID uint) (*model.Chef, error) {
	return s.chefRepo.GetWithMeals(chefID)
}

func (s *ChefService) VerifyChef(chefID uint) error {
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

func (s *ChefService) DeactivateChef(chefID uint) error {
	chef, err := s.chefRepo.GetByID(chefID)
	if err != nil {
		return err
	}
	if chef == nil {
		return errors.New("chef bulunamadı")
	}
	
	chef.IsActive = false
	return s.chefRepo.Update(chef)
}

func (s *ChefService) GetAllChefs() ([]model.Chef, error) {
	return s.chefRepo.GetAll()
}

func (s *ChefService) GetChefByID(chefID uint) (*model.Chef, error) {
	return s.chefRepo.GetByID(chefID)
}

func (s *ChefService) GetChefByUserID(userID uint) (*model.Chef, error) {
	return s.chefRepo.GetByUserID(userID)
}
