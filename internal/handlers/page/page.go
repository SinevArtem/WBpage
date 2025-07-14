package page

import (
	"encoding/json"
	"net/http"

	"github.com/SinevArtem/WBpage.git/internal/cache"
	"github.com/go-chi/chi"
)

func PageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		http.ServeFile(w, r, "static/templates/index.html")

	}
}

func GetOrderHandler(cache *cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderUID := chi.URLParam(r, "order_uid")
		if orderUID == "" {
			http.Error(w, `{"error": "order_uid is required"}`, http.StatusBadRequest)
			return
		}

		order, ok := cache.Get(orderUID)
		if !ok {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(order)
	}
}
