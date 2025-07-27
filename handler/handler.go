package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Adit0507/url-shortener/bloom"
	"github.com/Adit0507/url-shortener/snowflake"
	"github.com/Adit0507/url-shortener/store"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// converts num to base62 string
func ToBase62(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	chars := make([]byte, 0)
	for num > 0 {
		chars = append([]byte{base62Chars[num%62]}, chars...)
		num /= 62
	}

	return string(chars)
}

// handlin URL shortening
func ShortenHandler(snowFlakeGen *snowflake.Snowflake, bloomFilter *bloom.BloomFilter, urlStore *store.URLStore, taskChan chan<- store.Task) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := r.FormValue("url")
		if url == "" {
			http.Error(w, "URLL is required", http.StatusBadRequest)
			return
		}

		id := snowFlakeGen.Generate()
		shortCode := ToBase62(id)

		// checkin bloomfilter for potential collision
		if bloomFilter.MightContain(shortCode) {
			if _, exists := urlStore.Get(shortCode); exists {
				id = snowFlakeGen.Generate()
				shortCode = ToBase62(id)
			}
		}

		// queue storage task
		select {
		case taskChan <- store.Task{ShortCode: shortCode, URL: url}:
			fmt.Fprintf(w, "Short URL: http://localhost:8080/%s\n", shortCode)
		case <-ctx.Done():
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
			return
		}
	}
}
