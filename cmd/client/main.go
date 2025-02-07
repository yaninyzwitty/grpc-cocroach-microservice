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
	productRequest := &pb.CreateProductRequest{
		Name:          "Lemon Cake",
		Description:   "Rich and moist lemon cake topped with creamy frosting.",
		Price:         15.99,
		Category:      "Dessert",
		Tags:          []string{"cake", "dessert", "chocolate"},
		ProductState:  pb.ProductState_PERISHABLE,
		ProductStatus: pb.ProductStatus_IN_STOCK,
		Variation: &pb.CreateProductRequest_Food{
			Food: &pb.FoodVariation{
				Ingredients:  "Flour, Sugar, Cocoa Powder, Eggs, Butter",
				Calories:     350,
				IsVegetarian: true,
			},
		},
	}

	res, err := productClient.CreateProduct(ctx, productRequest)
	if err != nil {
		slog.Error("failed to create product", "error", err)
		os.Exit(1)
	}

	slog.Info("Product created successfully", "productID", res.Id, "productName", res.Name, "tags", res.Tags, "variations", res.Variation)

	getProductRes, err := productClient.GetProduct(ctx, &pb.GetProductRequest{
		Id: int64(229102406774849537),
	})
	if err != nil {
		slog.Error("failed to get product", "error", err)
		os.Exit(1)
	}

	slog.Info("Product retrieved successfully", "productName", getProductRes.Product.Name, "tags", getProductRes.Product.Tags, "variations", getProductRes.Product.Variation, "value", getProductRes.Product.ProductStatus)

}
