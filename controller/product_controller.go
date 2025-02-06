package controller

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pb"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/sonyflake"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type productController struct {
	pool *pgxpool.Pool
	// memcachedClient *database.MemcachedClient
	pb.UnimplementedProductServiceServer
}

// NewProductController returns an instance that implements pb.ProductServiceServer.
func NewProductController(pool *pgxpool.Pool) pb.ProductServiceServer {
	return &productController{
		pool: pool,
		// memcachedClient: memcachedClient,
	}
}

func (c *productController) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	productID, err := sonyflake.GenerateID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate product id: %v", err)
	}

	product := &pb.Product{
		Id:            int64(productID),
		Name:          req.Name,
		Description:   req.Description,
		Price:         req.Price,
		Category:      req.Category,
		Tags:          req.Tags,
		ProductState:  req.ProductState,
		ProductStatus: req.ProductStatus,
	}

	slog.Info(req.Name)
	switch v := req.Variation.(type) {
	case *pb.CreateProductRequest_Clothing:
		product.Variation = &pb.Product_Clothing{Clothing: v.Clothing}
	case *pb.CreateProductRequest_Electronics:
		product.Variation = &pb.Product_Electronics{Electronics: v.Electronics}
	case *pb.CreateProductRequest_Food:
		product.Variation = &pb.Product_Food{Food: v.Food}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid product variation type")
	}

	var createdAt, updatedAt time.Time

	query := `INSERT INTO products (id, name, description, price, category, tags,  product_state, product_status, variation) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at, updated_at`
	err = c.pool.QueryRow(ctx, query, int64(productID), product.Name, product.Description, product.Price, product.Category, product.Tags, product.ProductState, product.ProductStatus, product.Variation).Scan(&createdAt, &updatedAt)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	response := &pb.CreateProductResponse{
		Id:            product.Id,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Category:      product.Category,
		Tags:          product.Tags,
		CreatedAt:     timestamppb.New(createdAt),
		UpdatedAt:     timestamppb.New(updatedAt),
		ProductState:  product.ProductState,
		ProductStatus: product.ProductStatus,
	}
	switch v := req.Variation.(type) {
	case *pb.CreateProductRequest_Clothing:
		response.Variation = &pb.CreateProductResponse_Clothing{Clothing: v.Clothing}
	case *pb.CreateProductRequest_Electronics:
		response.Variation = &pb.CreateProductResponse_Electronics{Electronics: v.Electronics}
	case *pb.CreateProductRequest_Food:
		response.Variation = &pb.CreateProductResponse_Food{Food: v.Food}
	}

	return response, nil

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
