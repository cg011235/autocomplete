package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cg011235/autocomplete/internal/trie"
	"github.com/cg011235/autocomplete/pkg/models"
	"github.com/golang-jwt/jwt"
	"github.com/patrickmn/go-cache"
)

var (
	trieV1  = trie.NewTrie()
	cacheV1 = cache.New(5*time.Minute, 10*time.Minute)
)

var validCredentials = map[string]string{
	"user1": "password123",
}

var secretKey []byte

// SetSecretKey sets the JWT secret key.
func SetSecretKey(key []byte) {
	secretKey = key
}

// LoginHandler handles user login and issues a JWT token.
// @Summary Issue JWT token
// @Description Authenticates the user and issues a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body models.Credentials true "User credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/login [post]
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
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
// @Param words body models.AddWordsRequest true "List of words"
// @Success 200 {object} models.AddWordsResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/words [post]
func AddWordsHandlerV1(w http.ResponseWriter, r *http.Request) {
	var request models.AddWordsRequest
	json.NewDecoder(r.Body).Decode(&request)
	for _, word := range request.Words {
		trieV1.Insert(strings.ToLower(word))
		cacheV1.Flush() // Clear cache whenever new words are added
	}
	w.WriteHeader(http.StatusOK)
	response := models.AddWordsResponse{
		Status:  "success",
		Message: "Words added successfully.",
	}
	json.NewEncoder(w).Encode(response)
}

// ListWordsHandlerV1 retrieves words from the Trie based on the given prefix.
// @Summary Retrieve words and count
// @Description Retrieves all words stored in the Trie or looks up words that start with a given prefix, along with the total word count
// @Tags words
// @Accept json
// @Produce json
// @Param prefix query string false "Prefix to search for"
// @Success 200 {object} models.ListWordsResponse
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

	response := models.ListWordsResponse{
		Status: "success",
		Count:  count,
		Data:   results,
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
// @Param word body models.DeleteWordsRequest true "Word to delete"
// @Success 200 {object} models.DeleteWordsResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/words [delete]
func DeleteWordsHandlerV1(w http.ResponseWriter, r *http.Request) {
	var request models.DeleteWordsRequest
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
	response := models.DeleteWordsResponse{
		Status:  "success",
		Message: "Word(s) deleted successfully.",
	}
	json.NewEncoder(w).Encode(response)
}

// WordsExistsHandlerV1 checks if a word exists in the Trie.
// @Summary Check if a word exists in the Trie
// @Description Checks if a word exists in the Trie
// @Tags words
// @Accept json
// @Produce json
// @Param word query string true "Word to check"
// @Success 200 {object} models.CheckWordExistsResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/words/exists [get]
func WordsExistsHandlerV1(w http.ResponseWriter, r *http.Request) {
	word := r.URL.Query().Get("word")
	if word == "" {
		http.Error(w, "Missing 'word' query parameter", http.StatusBadRequest)
		return
	}

	exists := trieV1.Exists(word)

	response := models.CheckWordExistsResponse{
		Status: "success",
		Exists: exists,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
