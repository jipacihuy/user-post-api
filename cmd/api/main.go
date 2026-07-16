package main

import (
	"log"
	"net/http"
	"os"

	"user-post-api/internal/config"
	"user-post-api/internal/handler"
	"user-post-api/internal/middleware"
	"user-post-api/internal/redis"
	"user-post-api/internal/repository"
	"user-post-api/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	db := config.InitDB()
	rdb := config.InitRedis()

	cacheService := redis.NewCacheService(rdb)

	// ===== REPOSITORY =====
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)

	// ===== SERVICE =====
	userService := service.NewUserService(userRepo, cacheService)
	postService := service.NewPostService(postRepo, cacheService)

	// ===== HANDLER =====
	userHandler := handler.NewUserHandler(userService)
	postHandler := handler.NewPostHandler(postService)

	// ===== ROUTER =====
	r := mux.NewRouter()

	// ===== PUBLIC ROUTES =====
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	r.HandleFunc("/posts", postHandler.GetPosts).Methods("GET")
	r.HandleFunc("/posts/{id}", postHandler.GetPost).Methods("GET")

	// ===== PROTECTED ROUTES =====
	r.HandleFunc("/users", middleware.Auth(userHandler.GetUsers)).Methods("GET")
	r.HandleFunc("/users/{id}", middleware.Auth(userHandler.GetUser)).Methods("GET")

	r.HandleFunc("/posts", middleware.Auth(postHandler.CreatePost)).Methods("POST")
	r.HandleFunc("/posts/{id}", middleware.Auth(postHandler.DeletePost)).Methods("DELETE")

	// ===== SERVER =====
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}