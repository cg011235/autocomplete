// Package trie provides the implementation of a Trie data structure.
package trie

import "sync"

// Node represents a single node in the Trie.
type Node struct {
	Children map[rune]*Node
	IsWord   bool
}

// NewNode creates and returns a new Trie node.
func NewNode() *Node {
	return &Node{Children: make(map[rune]*Node), IsWord: false}
}

// Trie represents the Trie data structure with a root node and a mutex for concurrency control.
type Trie struct {
	Root *Node
	mu   sync.RWMutex
}

// NewTrie creates and returns a new Trie.
func NewTrie() *Trie {
	return &Trie{Root: NewNode()}
}

// Insert adds a word to the Trie.
func (t *Trie) Insert(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	node := t.Root
	for _, char := range word {
		if _, found := node.Children[char]; !found {
			node.Children[char] = NewNode()
		}
		node = node.Children[char]
	}
	node.IsWord = true
}

// Delete removes a word from the Trie.
func (t *Trie) Delete(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	node := t.Root
	stack := []*Node{node}
	for _, char := range word {
		if _, found := node.Children[char]; !found {
			return // Word not found
		}
		node = node.Children[char]
		stack = append(stack, node)
	}
	if !node.IsWord {
		return // Word not found
	}
	node.IsWord = false
	for i := len(word) - 1; i >= 0; i-- {
		char := rune(word[i])
		node := stack[i]
		child := stack[i+1]
		if len(child.Children) == 0 && !child.IsWord {
			delete(node.Children, char)
		}
	}
}

// Exists checks if a word exists in the Trie.
func (t *Trie) Exists(word string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	node := t.Root
	for _, char := range word {
		if _, found := node.Children[char]; !found {
			return false
		}
		node = node.Children[char]
	}
	return node.IsWord
}

// CollectWords collects all words in the Trie starting from the given node.
func (t *Trie) CollectWords(node *Node, prefix string) []string {
	var results []string
	if node.IsWord {
		results = append(results, prefix)
	}
	for char, child := range node.Children {
		results = append(results, t.CollectWords(child, prefix+string(char))...)
	}
	return results
}

// CountWords counts the total number of words in the Trie starting from the given node.
func (t *Trie) CountWords(node *Node) int {
	count := 0
	if node.IsWord {
		count++
	}
	for _, child := range node.Children {
		count += t.CountWords(child)
	}
	return count
}
