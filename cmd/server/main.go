package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cg011235/autocomplete/internal/handlers"
	"github.com/cg011235/autocomplete/internal/middleware"

	"github.com/gorilla/mux"
)

func main() {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Fatal("SECRET_KEY environment variable is required")
	}
	middleware.SetSecretKey([]byte(secretKey))

	r := mux.NewRouter()

	r.Use(middleware.JwtMiddleware)
	r.Use(middleware.RateLimitMiddleware)

	// Version 1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.HandleFunc("/add", handlers.AddWordsHandlerV1).Methods("POST")
	v1.HandleFunc("/lookup", handlers.LookupPrefixV1).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
