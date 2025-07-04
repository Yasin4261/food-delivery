package service

import (
	"errors"
	"ecommerce/internal/auth"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// UserService - kullanıcı iş mantığı
type UserService struct {
	userRepo        *repository.UserRepository
	jwtManager      *auth.JWTManager
	passwordManager *auth.PasswordManager
}

func NewUserService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		passwordManager: auth.NewPasswordManager(),
	}
}

func (s *UserService) Register(req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Email kontrolü - kullanıcı zaten var mı?
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("bu email adresi zaten kullanımda")
	}
	
	// Şifreyi hashle
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	
	// Yeni kullanıcı oluştur
	user := &model.User{
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "customer", // Varsayılan rol
	}
	
	// Veritabanına kaydet
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}
	
	// JWT token oluştur
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	
	return &model.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.AuthResponse, error) {
	// Email ile kullanıcı bul
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("email veya şifre hatalı")
	}
	
	// Şifre doğrulama
	if !s.passwordManager.CheckPasswordHash(req.Password, user.Password) {
		return nil, errors.New("email veya şifre hatalı")
	}
	
	// JWT token oluştur
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	
	return &model.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *UserService) GetProfile(userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *UserService) UpdateProfile(userID uint, req *model.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("kullanıcı bulunamadı")
	}
	
	// Güncelle
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}
