package main

import (
	"context"
	"log"
	"travel-api/internal/injector"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	db, err := pgxpool.New(context.Background(), "postgres://dev_user:dev_pass@localhost:5432/dev_db")
	if err != nil {
		log.Fatal("connection failed")
	}

	router := gin.Default()

	injector.NewTripHandler(db).RegisterAPI(router)

	router.Run()
}
