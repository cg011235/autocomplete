package trie

import "sync"

type Node struct {
	children map[rune]*Node
	isWord   bool
}

func NewNode() *Node {
	return &Node{children: make(map[rune]*Node), isWord: false}
}

type Trie struct {
	root *Node
	mu   sync.RWMutex
}

func NewTrie() *Trie {
	return &Trie{root: NewNode()}
}

func (t *Trie) Insert(word string) {
	if word == "" {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	node := t.root
	for _, char := range word {
		if _, found := node.children[char]; !found {
			node.children[char] = NewNode()
		}
		node = node.children[char]
	}

	node.isWord = true
}

func collectWords(node *Node, currentPrefix string) []string {
	var results []string
	if node.isWord {
		results = append(results, currentPrefix)
	}
	for ch, child := range node.children {
		results = append(results, collectWords(child, currentPrefix+string(ch))...)
	}
	return results
}

func (t *Trie) Search(prefix string) []string {
	if prefix == "" {
		return []string{}
	}

	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for _, ch := range prefix {
		if _, found := node.children[ch]; !found {
			return []string{} // prefix not found
		}
		node = node.children[ch]
	}

	return collectWords(node, prefix)
}
