// Package main is the entry point of the Trie-based autocomplete service.
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
	handlers.SetSecretKey([]byte(secretKey))

	r := mux.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RateLimitMiddleware)

	// Login route does not require JWT middleware
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")

	// Version 1 routes
	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.Use(middleware.JwtMiddleware)
	v1.HandleFunc("/", handlers.RootHandler).Methods("GET")
	v1.HandleFunc("/words", handlers.AddWordsHandlerV1).Methods("POST")
	v1.HandleFunc("/words", handlers.ListWordsHandlerV1).Methods("GET")
	v1.HandleFunc("/words", handlers.DeleteWordsHandlerV1).Methods("DELETE")
	v1.HandleFunc("/words/exists", handlers.WordsExistsHandlerV1).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
