package page

import (
	"encoding/json"
	"net/http"

	"github.com/SinevArtem/WBpage.git/internal/cache"
)

func PageHandler(cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders := cache.GetAll()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}
}
