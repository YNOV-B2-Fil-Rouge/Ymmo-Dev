// Command api is the entry point of the Ymmo back-end.
// Responsibilities: load config -> connect DB -> build router ->
// start HTTP server with graceful shutdown.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"ymmo/internal/config"
	"ymmo/internal/database"
	"ymmo/internal/router"
)

func main() {
	// 1. Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// 2. Database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	log.Println("connected to MariaDB")

	// 3. HTTP server
	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router.New(cfg, db),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the server in a goroutine so it does not block shutdown handling.
	go func() {
		log.Printf("ymmo-api listening on %s (%s)", srv.Addr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 4. Graceful shutdown on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Println("stopped cleanly")
}
