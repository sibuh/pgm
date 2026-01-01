package main

import (
	"database/sql"
	"log"
	"os"
	http "pgm/internal/handler/payment"
	rabbitmq "pgm/internal/queue"
	postgres "pgm/internal/repo"
	in "pgm/internal/repo/init"
	"pgm/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Database
	db, err := in.NewPostgresDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// RabbitMQ Publisher
	publisher, err := rabbitmq.NewRabbitMQPublisher()
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	// Note: In a real app, we'd handle closing the publisher gracefully

	// Repository
	repo := postgres.NewPaymentPostgresRepository(db)

	// UseCase
	uc := service.NewPaymentService(repo, publisher)

	// Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &http.CustomValidator{Validator: validator.New()}

	// Handlers
	http.NewPaymentHandler(e, uc)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}

func runMigrations(db *sql.DB) error {
	migrationFile := "migrations/000001_create_payments_table.up.sql"
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}

	log.Println("Migrations ran successfully")
	return nil
}
