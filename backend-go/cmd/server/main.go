// Command server runs the AI-chat Go backend: HTTP API, SSE chat stream and
// (optionally) the Telegram admin bot.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/api"
	"ai-chat-backend/internal/bot"
	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/service"
	"ai-chat-backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	for _, w := range cfg.ValidateRuntime() {
		fmt.Printf("[boot] %s\n", w)
	}

	// Make sure the SQLite parent directory exists (Docker volume mount points).
	if dir := filepath.Dir(cfg.SQLITEPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	st, err := store.Open(cfg.SQLITEPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "store open error: %v\n", err)
		os.Exit(1)
	}
	defer st.DB.Close()

	router := ai.NewRouter(cfg)
	svc := service.New(cfg, st, router)
	handler := api.New(svc, router)

	// Telegram admin bot: starts only when TG_BOT_TOKEN is configured.
	botCtx, botCancel := context.WithCancel(context.Background())
	defer botCancel()
	bot.Start(botCtx, cfg, svc)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 15 * time.Second,
		ReadTimeout:       60 * time.Second, // kills slow-body drip attacks; SSE bodies are empty
		IdleTimeout:       120 * time.Second,
		// No WriteTimeout: SSE streams legitimately run for hours (nginx
		// proxy_read_timeout is 3600s) and WriteTimeout would sever them.
	}

	go func() {
		fmt.Printf("backend server running at http://localhost:%s (AI_MODE=%s, models=%d)\n", cfg.Port, cfg.AIMode, len(cfg.AIProviders))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "listen error: %v\n", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("[shutdown] signal received, draining...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	botCancel()
	fmt.Println("[shutdown] bye")
}
