package service

import (
	"errors"
	"ecommerce/internal/model"
	"ecommerce/internal/repository"
)

// ProductService - ürün iş mantığı
type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetAllProducts() ([]model.Product, error) {
	products, err := s.productRepo.GetAll()
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *ProductService) GetProduct(id uint) (*model.Product, error) {
	if id == 0 {
		return nil, errors.New("geçersiz ürün ID")
	}
	
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("ürün bulunamadı")
	}
	
	return product, nil
}

func (s *ProductService) CreateProduct(req *model.CreateProductRequest) (*model.Product, error) {
	// Validasyonlar
	if req.Name == "" {
		return nil, errors.New("ürün adı boş olamaz")
	}
	if req.Price <= 0 {
		return nil, errors.New("ürün fiyatı 0'dan büyük olmalıdır")
	}
	if req.Stock < 0 {
		return nil, errors.New("stok miktarı negatif olamaz")
	}
	
	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
	}
	
	err := s.productRepo.Create(product)
	if err != nil {
		return nil, err
	}
	
	return product, nil
}

func (s *ProductService) UpdateProduct(id uint, req *model.UpdateProductRequest) (*model.Product, error) {
	if id == 0 {
		return nil, errors.New("geçersiz ürün ID")
	}
	
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("ürün bulunamadı")
	}
	
	// Güncelle
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.CategoryID > 0 {
		product.CategoryID = req.CategoryID
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	
	err = s.productRepo.Update(product)
	if err != nil {
		return nil, err
	}
	
	return product, nil
}

func (s *ProductService) DeleteProduct(id uint) error {
	if id == 0 {
		return errors.New("geçersiz ürün ID")
	}
	
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("ürün bulunamadı")
	}
	
	return s.productRepo.Delete(id)
}

func (s *ProductService) GetProductsByCategory(categoryID uint) ([]model.Product, error) {
	if categoryID == 0 {
		return nil, errors.New("geçersiz kategori ID")
	}
	
	return s.productRepo.GetByCategory(categoryID)
}

func (s *ProductService) SearchProducts(query string) ([]model.Product, error) {
	if query == "" {
		return s.GetAllProducts()
	}
	
	return s.productRepo.Search(query)
}
