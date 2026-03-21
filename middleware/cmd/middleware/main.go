package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"middleware/consumer"
	"middleware/data"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	dbDSN := buildPostgresDSN()
	rabbitDSN := buildRabbitMQDSN()

	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		log.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Error("failed to reach database", "error", err)
		os.Exit(1)
	}
	log.Info("database connected")

	repo := data.NewSensorRepository(db)

	job := consumer.NewJob(repo, rabbitDSN, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	job.Run(ctx)
	log.Info("middleware exited cleanly")
}

func buildPostgresDSN() string {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "admin")
	pass := getEnv("DB_PASSWORD", "admin")
	name := getEnv("DB_NAME", "database")
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)
}

func buildRabbitMQDSN() string {
	host := getEnv("RABBITMQ_HOST", "localhost")
	port := getEnv("RABBITMQ_PORT", "5672")
	user := getEnv("RABBITMQ_USER", "admin")
	pass := getEnv("RABBITMQ_PASSWORD", "admin")
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}