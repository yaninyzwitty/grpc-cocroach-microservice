package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/yaninyzwitty/grpc-cocroach-microservice/pb"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pkg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	address := fmt.Sprintf(":%d", cfg.Server.Port)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to create client", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	productClient := pb.NewProductServiceClient(conn)

	getProductRes, err := productClient.GetProduct(ctx, &pb.GetProductRequest{
		Id: int64(229577284481220609),
	})
	if err != nil {
		slog.Error("failed to get product", "error", err)
		os.Exit(1)
	}

	slog.Info("here", "val", getProductRes)

}

// grpcurl -d "{\"id\": 229577284481220609}" -proto proto\products.proto -import-path ./ -plaintext localhost:50051 products.ProductService/GetProduct
