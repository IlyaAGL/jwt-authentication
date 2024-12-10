package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	routes "github.com/agl/auth/internal/routes"
	"github.com/jackc/pgx/v5"
)

const (
	DATABASE_URL = "postgres://postgres:postgres@localhost:6432/postgres"
)

func main() {
	conn, err := pgx.Connect(context.Background(), DATABASE_URL)

	if err != nil {
		log.Fatalf("Cant connect to db: %v", err)
	}
	defer conn.Close(context.Background())

	fmt.Println("DB WAS CONNECTED")
	routes.SetupRoutes(conn)

	http.ListenAndServe(":5050", nil)
}
