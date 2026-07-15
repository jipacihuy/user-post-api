package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	initDB()

	router := mux.NewRouter()

	// Routes...
	router.HandleFunc("/register", Register).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")
	router.HandleFunc("/profile", authMiddleware(GetProfile)).Methods("GET")
	// ... rest of routes
	// User routes
	router.HandleFunc("/users", GetUsers).Methods("GET")
	router.HandleFunc("/users/{id}", GetUser).Methods("GET")

	// Post routes
	router.HandleFunc("/posts", GetPosts).Methods("GET")
	router.HandleFunc("/posts/{id}", GetPost).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}