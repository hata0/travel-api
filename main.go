package main

import (
	"context"
	"log"
	"travel-api/infrastructure/database"
	"travel-api/injector"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbConnection, err := pgxpool.New(context.Background(), "postgres://dev_user:dev_pass@localhost:5432/dev_db")
	if err != nil {
		log.Fatal("connection failed")
	}
	queries := database.New(dbConnection)

	router := gin.Default()

	tripController := injector.NewTripController(queries)
	tripController.Register(router)

	router.Run()
}
