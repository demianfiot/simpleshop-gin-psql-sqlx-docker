package service

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"prac/pkg/repository"
	"prac/todo"
	"time"
)

type UserService struct {
	repo         repository.User
	cacheRepo    repository.CacheRepository
	cacheTTL     time.Duration
	listCacheTTL time.Duration
}

func NewUserService(repo repository.User, cacheRepo repository.CacheRepository) *UserService {
	return &UserService{
		repo:         repo,
		cacheRepo:    cacheRepo,
		cacheTTL:     10 * time.Minute, // TTL user
		listCacheTTL: 5 * time.Minute,  // TTL list
	}
}

func (s *UserService) CreateUser(ctx context.Context, input todo.User) (int, error) {
	hashedPassword := s.generatePasswordHash(input.PasswordHash)

	user := todo.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         input.Role,
	}

	userID, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	// -cache
	s.cacheRepo.DeleteByPattern(ctx, "users:*")

	return userID, nil
}

// cache keys
func (s *UserService) userCacheKey(id uint) string {
	return fmt.Sprintf("user:%d", id)
}

func (s *UserService) usersListCacheKey() string {
	return "users:list"
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]todo.User, error) {

	cacheKey := s.usersListCacheKey()

	// cache
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var users []todo.User
		if err := json.Unmarshal(cachedData, &users); err == nil {
			return users, nil
		}
		return users, nil
	}

	//db
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		s.cacheRepo.Set(ctx, cacheKey, users, s.listCacheTTL)
	}

	return users, nil
}

func (s *UserService) GetUserByID(ctx context.Context, userID uint) (todo.User, error) {
	cacheKey := s.userCacheKey(userID)
	// cache
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var user todo.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			return user, nil
		}
		return user, nil
	}

	//db
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return todo.User{}, err
	}

	s.cacheRepo.Set(ctx, cacheKey, user, s.cacheTTL)

	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (todo.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)
	// cache
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var user todo.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			return user, nil
		}
		return user, nil
	}
	//db
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return todo.User{}, err
	}

	s.cacheRepo.Set(ctx, cacheKey, user, s.cacheTTL)
	s.cacheRepo.Set(ctx, s.userCacheKey(user.ID), user, s.cacheTTL)

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, input todo.UpdateUserInput) (todo.UpdateUserInput, error) {

	result, err := s.repo.UpdateUser(ctx, userID, input)
	if err != nil {
		return result, err
	}

	// -cache
	cacheKey := s.userCacheKey(userID)
	s.cacheRepo.Delete(ctx, cacheKey)
	s.cacheRepo.Delete(ctx, "user:email:*")
	s.cacheRepo.DeleteByPattern(ctx, "users:list*")

	return result, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	// -cache
	cacheKey := s.userCacheKey(uint(id))
	s.cacheRepo.Delete(ctx, cacheKey)
	s.cacheRepo.DeleteByPattern(ctx, "users:*")

	return nil
}
func (s *UserService) generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
