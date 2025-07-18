package main

import (
	"context"
	"log"
	"travel-api/internal/injector"

	"travel-api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn, err := config.DSN()
	if err != nil {
		log.Fatal(err)
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("connection failed")
	}

	router := gin.Default()

	injector.NewTripHandler(db).RegisterAPI(router)

	router.Run()
}
