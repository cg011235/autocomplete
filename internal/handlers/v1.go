// Package handlers provides the API request handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cg011235/autocomplete/internal/trie"
	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
)

var (
	trieV1  = trie.NewTrie()
	cacheV1 = cache.New(5*time.Minute, 10*time.Minute)
)

// validCredentials contains the mock username and password for authentication.
var validCredentials = map[string]string{
	"user1": "password123",
}

// secretKey holds the JWT secret key.
var secretKey []byte

// SetSecretKey sets the JWT secret key.
func SetSecretKey(key []byte) {
	secretKey = key
}

// Credentials represents the JSON payload for login requests.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler handles user login and issues a JWT token.
// @Summary Issue JWT token
// @Description Authenticates the user and issues a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body handlers.Credentials true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if password, ok := validCredentials[creds.Username]; !ok || password != creds.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": creds.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"token": tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RootHandler provides an overview of the API, including available endpoints and their descriptions.
// @Summary Root endpoint
// @Description Provides an overview of the API, including available endpoints and their descriptions
// @Tags root
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router / [get]
func RootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Welcome to the Trie-based Autocomplete API",
		"endpoints": []map[string]string{
			{"method": "POST", "endpoint": "/api/login", "description": "Authenticate, generate token"},
			{"method": "POST", "endpoint": "/api/v1/words", "description": "Add words to the Trie"},
			{"method": "GET", "endpoint": "/api/v1/words", "description": "Lookup words that start with a given prefix or retrieve all words"},
			{"method": "DELETE", "endpoint": "/api/v1/words", "description": "Delete a word from the Trie or clear all words"},
			{"method": "GET", "endpoint": "/api/v1/words/exists", "description": "Check if a word exists in the Trie"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddWordsHandlerV1 adds words to the Trie.
// @Summary Add words to the Trie
// @Description Adds words to the Trie
// @Tags words
// @Accept json
// @Produce json
// @Param words body object true "List of words"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/words [post]
func AddWordsHandlerV1(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Words []string `json:"words"`
	}
	json.NewDecoder(r.Body).Decode(&request)
	for _, word := range request.Words {
		trieV1.Insert(strings.ToLower(word))
		cacheV1.Flush() // Clear cache whenever new words are added
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Words added successfully."})
}

// ListWordsHandlerV1 retrieves words from the Trie based on the given prefix.
// @Summary Retrieve words and count
// @Description Retrieves all words stored in the Trie or looks up words that start with a given prefix, along with the total word count
// @Tags words
// @Accept json
// @Produce json
// @Param prefix query string false "Prefix to search for"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/words [get]
func ListWordsHandlerV1(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	var results []string
	var count int

	if cachedResult, found := cacheV1.Get(prefix); found {
		results = cachedResult.([]string)
		count = len(results)
	} else {
		if prefix == "" {
			results = trieV1.CollectWords(trieV1.Root, "")
			count = trieV1.CountWords(trieV1.Root)
		} else {
			node := trieV1.Root
			for _, char := range prefix {
				if _, found := node.Children[char]; !found {
					results = []string{}
					count = 0
					break
				}
				node = node.Children[char]
			}
			if count == 0 {
				results = trieV1.CollectWords(node, prefix)
				count = len(results)
			}
		}
		cacheV1.Set(prefix, results, cache.DefaultExpiration)
	}

	response := map[string]interface{}{
		"status": "success",
		"count":  count,
		"data":   results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteWordsHandlerV1 deletes words from the Trie based on the given request.
// @Summary Delete words from the Trie
// @Description Deletes a word from the Trie if the request contains a word, otherwise clears all words
// @Tags words
// @Accept json
// @Produce json
// @Param word body object true "Word to delete"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/words [delete]
func DeleteWordsHandlerV1(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Word string `json:"word"`
	}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.Word == "" {
		trieV1 = trie.NewTrie() // Clear all words
		cacheV1.Flush()         // Clear cache
	} else {
		trieV1.Delete(request.Word)
		cacheV1.Delete(request.Word) // Remove from cache
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Word(s) deleted successfully."})
}

// WordsExistsHandlerV1 checks if a word exists in the Trie.
// @Summary Check if a word exists in the Trie
// @Description Checks if a word exists in the Trie
// @Tags words
// @Accept json
// @Produce json
// @Param word query string true "Word to check"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/words/exists [get]
func WordsExistsHandlerV1(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "Missing 'word' query parameter", http.StatusBadRequest)
		return
	}

	exists := trieV1.Exists(word)

	response := map[string]interface{}{
		"status": "success",
		"exists": exists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
