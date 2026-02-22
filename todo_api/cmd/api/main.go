package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"todo/internal/handlers"
	"todo/internal/middleware"
	"todo/internal/repository"
)

func main() {
	repo := repository.NewTaskRepo()
	taskHandler := &handlers.TaskHandler{Repo: repo}

	mux := http.NewServeMux()
	mux.Handle("/tasks", taskHandler)

	finalHandler := middleware.Logging(middleware.Auth(mux))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: finalHandler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Server is running on http://localhost:8080\n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-done
	fmt.Println("\nServer is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited properly")
}
