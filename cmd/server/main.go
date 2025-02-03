package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/database"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/helpers"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pkg"
)

func main() {

	var cfg pkg.Config

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml", "error", err)
		os.Exit(1)
	}
	defer file.Close()

	if err := cfg.LoadFile(file); err != nil {
		slog.Error("failed to load config.yaml", "error", err)
		os.Exit(1)
	}

	err = godotenv.Load()
	if err != nil {
		slog.Error("failed to load .env file", "error", err)
		os.Exit(1)
	}

	COCROACH_DB_PASSWORD := helpers.GetEnvOrDefault("COCROACH_DB_PASSWORD", "")
	COCROACH_USERNAME := helpers.GetEnvOrDefault("COCROACH_USERNAME", "")

	slog.Info("Environment variables loaded successfully", "COCROACH_DB_PASSWORD", COCROACH_DB_PASSWORD != "", "COCROACH_USERNAME", COCROACH_USERNAME != "")

	dbConfig := database.DbConfig{
		Host:     cfg.Database.Hostname,
		Port:     cfg.Database.Port,
		User:     COCROACH_USERNAME,
		Password: COCROACH_DB_PASSWORD,
		DbName:   cfg.Database.Database,
		SSLMode:  cfg.Database.SSLMode,
		MaxConn:  500,
	}
	pool, err := dbConfig.NewPgxPool(ctx, 30)
	if err != nil {
		slog.Error("failed to create pgx pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := dbConfig.Ping(ctx, pool, 30); err != nil {
		slog.Error("failed to ping db", "error", err)
		os.Exit(1)
	}

}
