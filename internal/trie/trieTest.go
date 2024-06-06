package trie

import (
	"testing"
)

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func TestInsertAndSearch(t *testing.T) {
	trie := NewTrie()

	// Test inserting an empty string and searching for it
	trie.Insert("")
	results := trie.Search("")
	if len(results) > 0 {
		t.Fatal("Invalid results for empty prefix")
	}

	// Insert some strings
	trie.Insert("magic")
	trie.Insert("magnet")
	trie.Insert("maggie")
	trie.Insert("maggot")
	trie.Insert("ma")
	trie.Insert("megan")
	trie.Insert("mama")
	trie.Insert("mam")

	// Search valid prefix
	results = trie.Search("mag")
	expectedResults := []string{"magic", "magnet", "maggie", "maggot"}
	for _, expected := range expectedResults {
		if !contains(results, expected) {
			t.Fatalf("Expected result '%s' not found for prefix 'mag'", expected)
		}
	}

	// Ensure no extra results are included
	if len(results) != len(expectedResults) {
		t.Fatalf("Unexpected results for prefix 'mag': %v", results)
	}

	// Search invalid prefix
	results = trie.Search("a")
	if len(results) > 0 {
		t.Fatal("Results should be empty for un-inserted search")
	}

	// Search valid prefix with single character
	results = trie.Search("ma")
	expectedResults = []string{"magic", "magnet", "maggie", "maggot", "ma", "mama", "mam"}
	for _, expected := range expectedResults {
		if !contains(results, expected) {
			t.Fatalf("Expected result '%s' not found for prefix 'ma'", expected)
		}
	}

	// Ensure no extra results are included
	if len(results) != len(expectedResults) {
		t.Fatalf("Unexpected results for prefix 'ma': %v", results)
	}
}
