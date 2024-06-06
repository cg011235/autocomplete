package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cg011235/autocomplete/internal/trie"
	"github.com/patrickmn/go-cache"
)

var (
	trieV1  = trie.NewTrie()
	cacheV1 = cache.New(5*time.Minute, 10*time.Minute)
)

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

func LookupPrefixV1(w http.ResponseWriter, r *http.Request) {
	prefix := strings.ToLower(r.URL.Query().Get("prefix"))
	if prefix == "" {
		http.Error(w, "Prefix is required", http.StatusBadRequest)
		return
	}

	// Get limit and offset from query parameters
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 10 // default limit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0 // default offset
	}

	// Get filter from query parameters
	filter := strings.ToLower(r.URL.Query().Get("filter"))

	// Check cache first
	cacheKey := prefix + strconv.Itoa(limit) + strconv.Itoa(offset) + filter
	if suggestions, found := cacheV1.Get(cacheKey); found {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "suggestions": suggestions})
		return
	}

	// If not found in cache, search in trie
	allSuggestions := trieV1.Search(prefix)

	// Apply filter
	var filteredSuggestions []string
	if filter != "" {
		for _, suggestion := range allSuggestions {
			if strings.Contains(suggestion, filter) {
				filteredSuggestions = append(filteredSuggestions, suggestion)
			}
		}
	} else {
		filteredSuggestions = allSuggestions
	}

	// Sort the filtered suggestions alphabetically
	sort.Strings(filteredSuggestions)

	// Apply pagination
	start := offset
	if start > len(filteredSuggestions) {
		start = len(filteredSuggestions)
	}
	end := start + limit
	if end > len(filteredSuggestions) {
		end = len(filteredSuggestions)
	}
	paginatedSuggestions := filteredSuggestions[start:end]

	// Store result in cache
	cacheV1.Set(cacheKey, paginatedSuggestions, cache.DefaultExpiration)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "suggestions": paginatedSuggestions})
}
