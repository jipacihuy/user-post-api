package service

import (
	"time"
	"user-post-api/internal/model"
	"user-post-api/internal/redis"
	"user-post-api/internal/repository"
)

type PostService struct {
	Repo  *repository.PostRepository
	Cache *redis.CacheService
}

func NewPostService(repo *repository.PostRepository, cache *redis.CacheService) *PostService {
	return &PostService{
		Repo:  repo,
		Cache: cache,
	}
}

func (s *PostService) GetAllPosts() ([]model.Post, error) {
	var posts []model.Post
	err := s.Cache.Get("posts:all", &posts)
	if err == nil {
		return posts, nil
	}
	posts, err = s.Repo.GetAll()
	if err != nil {
		return nil, err
	}
	_ = s.Cache.Set("posts:all", posts, 5*time.Minute)
	return posts, nil
}

func (s *PostService) GetPostByID(id int) (model.Post, error) {
	var post model.Post
	err := s.Cache.Get("post:"+string(rune(id)), &post)
	if err == nil {
		return post, nil
	}
	post, err = s.Repo.GetByID(id)
	if err != nil {
		return post, err
	}
	_ = s.Cache.Set("post:"+string(rune(id)), post, 10*time.Minute)
	return post, nil
}

func (s *PostService) CreatePost(post model.Post) (int, error) {
	id, err := s.Repo.Create(post)
	if err == nil {
		_ = s.Cache.Delete("posts:*")
	}
	return id, err
}

func (s *PostService) DeletePost(id int) error {
	err := s.Repo.Delete(id)
	if err == nil {
		_ = s.Cache.Delete("posts:*")
		_ = s.Cache.Delete("post:*")
	}
	return err
}