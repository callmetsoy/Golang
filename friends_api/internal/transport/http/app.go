package http

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	sqlStorage "github.com/callmetsoy/user_api/internal/storage/sql"
	"github.com/callmetsoy/user_api/internal/transport/http/handlers"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type AppServer struct {
	server *http.Server
	db     *sqlx.DB
}

func NewAppServer(addr string, dbURL string) *AppServer {
	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		panic(err)
	}

	sqlStorage.NewUserRepo(db)

	userHandler := handlers.NewUserHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/users", userHandler.GetUsers)
	mux.HandleFunc("/common-friends", userHandler.GetCommonFriends)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return &AppServer{server: srv, db: db}
}

func (a *AppServer) Start() {
	go func() {
		fmt.Println("Server starting on", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server error:", err)
		}
	}()
}

func (a *AppServer) Stop(ctx context.Context) error {
	fmt.Println("Server shutting down...")
	if err := a.db.Close(); err != nil {
		log.Println("DB close error:", err)
	}
	return a.server.Shutdown(ctx)
}
