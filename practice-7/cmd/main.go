package main

import (
    "log"
    "practice-7/internal/controller/http/v1"
    "practice-7/internal/entity"
    "practice-7/internal/usecase"
    "practice-7/internal/usecase/repo"
    "practice-7/pkg/logger"
    "practice-7/pkg/postgres"
    "practice-7/utils"
    "time"

    "github.com/gin-gonic/gin"
)

func main() {

    l := logger.New()

    pg, err := postgres.New()
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    err = pg.Conn.AutoMigrate(&entity.User{})
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    userRepo := repo.NewUserRepo(pg)

    userUseCase := usecase.NewUserUseCase(userRepo)

    rl := utils.NewRateLimiter(10, time.Minute*1) 

    r := gin.Default()

    api := r.Group("/api/v1")
    v1.NewUserRoutes(api, userUseCase, l, rl)

    log.Println("Server starting on :8080")
    if err := r.Run(":8080"); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}