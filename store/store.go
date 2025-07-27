package store

import "sync"

// storing short code to url mappings
type URLStore struct {
	store map[string]string
	mu    sync.RWMutex
}

// represents a storage task for the worker pool
type Task struct {
	ShortCode string
	URL       string
}

func NewURLStore() *URLStore {
	return &URLStore{store: make(map[string]string)}
}

// saves a url mapping
func (us *URLStore) Store(shortCode, url string) {
	us.mu.Lock()
	defer us.mu.Unlock()

	us.store[shortCode] = url
}

func (us *URLStore) Get(shortCode string) (string, bool) {
	us.mu.Lock()
	defer us.mu.Unlock()

	url, exists := us.store[shortCode]

	return url, exists
}
