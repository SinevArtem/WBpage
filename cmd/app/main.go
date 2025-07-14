package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/SinevArtem/WBpage.git/internal/cache"
	"github.com/SinevArtem/WBpage.git/internal/handlers/page"
	"github.com/SinevArtem/WBpage.git/internal/kafka"
	"github.com/SinevArtem/WBpage.git/internal/postgres"
	"github.com/go-chi/chi"
)

func main() {

	db, err := postgres.New()
	if err != nil {
		log.Printf("failed to init storage: %v", err)
		os.Exit(1)
	}

	cache := cache.New()

	// Восстанавливаем кэш из БД
	orders, err := db.LoadOrders()
	if err != nil {
		log.Printf("failed to restore cache: %v", err)
	}

	for _, order := range orders {
		cache.Set(order)
	}

	var wg sync.WaitGroup

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	router := chi.NewRouter()

	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.Get("/", page.PageHandler())
	router.Get("/order/{order_uid}", page.GetOrderHandler(cache))

	server := &http.Server{
		Addr:    ":3002",
		Handler: router,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP error: %v", err)
			done <- syscall.SIGTERM

		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := kafka.Customer(db.DB, cache, ctx); err != nil {
			log.Printf("%v", err)

		}

	}()

	<-done

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	wg.Wait()

	db.Close()

}
