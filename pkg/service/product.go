package service

import (
	"context"
	"encoding/json"
	"fmt"
	"prac/pkg/repository"
	"prac/todo"
	"time"
)

type ProductService struct {
	repo         repository.Product
	cacheRepo    repository.CacheRepository
	cacheTTL     time.Duration
	listCacheTTL time.Duration
}

func NewProductService(repo repository.Product, cacheRepo repository.CacheRepository) *ProductService {
	return &ProductService{
		repo:         repo,
		cacheRepo:    cacheRepo,
		cacheTTL:     10 * time.Minute, // TTL user
		listCacheTTL: 5 * time.Minute,  // TTL list
	}
}

// cache keys
func (s *ProductService) productCacheKey(id uint) string {
	return fmt.Sprintf("product:%d", id)
}

func (s *ProductService) productsListCacheKey() string {
	return "products:list"
}

func (s *ProductService) CreateProduct(ctx context.Context, input todo.CreateProductInput, sellerID uint) (int, error) {
	product := todo.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
		SellerID:    sellerID,
	}

	return s.repo.CreateProduct(ctx, product)
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]todo.Product, error) {
	cacheKey := s.productsListCacheKey()

	// cache
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var products []todo.Product
		if err := json.Unmarshal(cachedData, &products); err == nil {
			return products, nil
		}
		return products, nil
	}

	//db
	products, err := s.repo.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}

	if len(products) > 0 {
		s.cacheRepo.Set(ctx, cacheKey, products, s.listCacheTTL)
	}

	return products, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, productID uint) (todo.Product, error) {
	cacheKey := s.productCacheKey(productID)
	// cache
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var product todo.Product
		json.Unmarshal(cachedData, &product)
		return product, nil
	}

	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return todo.Product{}, err
	}

	s.cacheRepo.Set(ctx, cacheKey, product, s.cacheTTL)

	return product, nil

}
func (s *ProductService) UpdateProduct(ctx context.Context, productID uint, input todo.UpdateProductInput, sellerID uint) (todo.Product, error) {
	product := todo.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
		SellerID:    sellerID,
	}

	return s.repo.UpdateProduct(ctx, productID, product)
}
func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.DeleteProduct(ctx, id)
}
