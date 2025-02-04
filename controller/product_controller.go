package controller

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/database"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pb"
)

type productController struct {
	pool            *pgxpool.Pool
	memcachedClient *database.MemcachedClient
	pb.UnimplementedProductServiceServer
}

// NewProductController returns an instance that implements pb.ProductServiceServer.
func NewProductController(pool *pgxpool.Pool, memcachedClient *database.MemcachedClient) pb.ProductServiceServer {
	return &productController{
		pool:            pool,
		memcachedClient: memcachedClient,
	}
}

func (c *productController) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	return &pb.CreateProductResponse{}, nil
}

func (c *productController) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	return &pb.UpdateProductResponse{}, nil
}

func (c *productController) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	return &pb.DeleteProductResponse{}, nil
}

func (c *productController) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	return &pb.GetProductResponse{}, nil
}

func (c *productController) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	return &pb.ListProductsResponse{}, nil
}
