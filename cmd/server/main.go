package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/controller"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/database"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/helpers"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pb"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pkg"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/sonyflake"
	"google.golang.org/grpc"
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

	// initalize memcached client
	memcachedClient, err := database.NewMemcachedClient(cfg.Memcache.Host, cfg.Memcache.Port)
	if err != nil {
		slog.Error("failed to create memcached client", "error", err)
		os.Exit(1)
	}

	err = sonyflake.InitSonyFlake()
	if err != nil {
		slog.Error("failed to initialize sonyflake", "error", err)
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		slog.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	productController := controller.NewProductController(pool, memcachedClient)
	server := grpc.NewServer()

	pb.RegisterProductServiceServer(server, productController)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		slog.Info("Shutting down gRPC server...")

		// Gracefully stop the gRPC server
		server.GracefulStop()
		cancel()

		slog.Info("gRPC server has been stopped gracefully")
	}()

	slog.Info("Starting gRPC server", "port", cfg.Server.Port)
	if err := server.Serve(lis); err != nil {
		slog.Error("gRPC server encountered an error while serving", "error", err)
		os.Exit(1)
	}

}
