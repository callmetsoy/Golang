package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/callmetsoy/user_api/internal/transport/http"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	app := http.NewAppServer(":8080", dbURL)
	app.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Stop(ctx); err != nil {
		log.Println("Shutdown error:", err)
	}

	fmt.Println("Server stopped")
}
