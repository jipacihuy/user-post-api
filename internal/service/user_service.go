package service

import (
	"errors"
	"time"
	"log"
	"user-post-api/internal/model"
	"user-post-api/internal/redis"
	"user-post-api/internal/repository"
)

type UserService struct {
	Repo  *repository.UserRepository
	Cache *redis.CacheService
}

func NewUserService(repo *repository.UserRepository, cache *redis.CacheService) *UserService {
	return &UserService{
		Repo:  repo,
		Cache: cache,
	}
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	// Coba ambil dari cache
	var users []model.User
	err := s.Cache.Get("users:all", &users)
	if err == nil {
		log.Println("✅ Cache HIT for users:all")
		return users, nil
	}
	log.Println("❌ Cache MISS for users:all")

	// Cache miss, ambil dari DB
	users, err = s.Repo.GetAll()
	if err != nil {
		return nil, err
	}

	// Simpan ke cache (5 menit)
	_ = s.Cache.Set("users:all", users, 5*time.Minute)
	return users, nil
}

func (s *UserService) GetUserByID(id int) (model.User, error) {
	return s.Repo.GetByID(id)
}

func (s *UserService) GetUserByEmail(email string) (model.User, error) {
	return s.Repo.GetByEmail(email)
}

func (s *UserService) CreateUser(user model.User) (int, error) {
	existing, _ := s.Repo.GetByEmail(user.Email)
	if existing.ID != 0 {
		return 0, errors.New("email already exists")
	}
	return s.Repo.Create(user)
}

func (s *UserService) UpdateUser(id int, user model.User) error {
	return s.Repo.Update(id, user)
}

func (s *UserService) DeleteUser(id int) error {
	return s.Repo.Delete(id)
}
