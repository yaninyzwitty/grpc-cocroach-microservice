package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/helpers"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pkg"
)

func main() {
	var cfg pkg.Config
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

}
