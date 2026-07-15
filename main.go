package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	initDB()
	r := mux.NewRouter()

	// ... routes (sama seperti sebelumnya)

	// Ambil port dari environment (Railway kasih PORT otomatis)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}