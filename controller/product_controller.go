package controller

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/pb"
	"github.com/yaninyzwitty/grpc-cocroach-microservice/sonyflake"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
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
	if req.GetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "product id is required")
	}

	productID := req.GetId()
	name := req.GetName()
	description := req.GetDescription()
	price := req.GetPrice()
	category := req.GetCategory()
	tags := req.GetTags()
	productState := req.GetProductState()
	productStatus := req.GetProductStatus()
	variation := req.GetVariation()

	query := `
	UPDATE products
	SET 
		name = $2,
		description = $3,
		price = $4,
		category = $5,
		tags = $6,
		product_state = $7,
		product_status = $8,
		variation = $9,
		updated_at = $10
	WHERE id = $1
	RETURNING id, name, description, price, category, tags, product_state, product_status, variation, created_at, updated_at;
	`

	updatedAt := time.Now()
	var createdTime time.Time
	var updatedTime time.Time
	err := c.pool.QueryRow(ctx, query,
		productID, name, description, price, category, tags, productState, productStatus, variation, updatedAt).Scan(
		&productID, &name, &description, &price, &category, &tags, &productState, &productStatus, &variation, &createdTime, &updatedTime,
	)

	if err != nil {
		if err.Error() == "no rows in result set" { // Check for "not found" explicitly
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	product := &pb.Product{
		Id:            productID,
		Name:          name,
		Description:   description,
		Price:         price,
		Category:      category,
		Tags:          tags,
		ProductState:  productState,
		ProductStatus: productStatus,
		CreatedAt:     timestamppb.New(createdTime),
		UpdatedAt:     timestamppb.New(updatedTime),
	}

	// Handle the variation field
	switch v := req.Variation.(type) {
	case *pb.UpdateProductRequest_Clothing:
		product.Variation = &pb.Product_Clothing{Clothing: v.Clothing}
	case *pb.UpdateProductRequest_Electronics:
		product.Variation = &pb.Product_Electronics{Electronics: v.Electronics}
	case *pb.UpdateProductRequest_Food:
		product.Variation = &pb.Product_Food{Food: v.Food}
	}

	return &pb.UpdateProductResponse{
		Product: product,
	}, nil
}
func (c *productController) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	if req.GetProductId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "product id is required")
	}
	productID := req.GetProductId()
	query := `
	DELETE FROM products
	WHERE id = $1
`
	_, err := c.pool.Exec(ctx, query, productID) // Use Exec for DELETE
	if err != nil {
		// Check if the product exists before deletion
		var count int
		countQuery := `SELECT COUNT(*) FROM products WHERE id = $1`
		errCount := c.pool.QueryRow(ctx, countQuery, productID).Scan(&count)
		if errCount == nil && count == 0 {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	return &pb.DeleteProductResponse{
		Deleted: true,
	}, nil
}

func (c *productController) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	// Validate the request: ensure the product ID is provided.
	if req.GetId() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "product id is required")
	}

	// Define the SQL query to retrieve the product.
	query := `
		SELECT id, name, description, price, category, tags, created_at, updated_at, product_state, product_status, variation
		FROM products 
		WHERE id = $1
	`

	var product pb.Product
	var variationData []byte
	var createdAt, updatedAt time.Time
	// Scan product_state and product_status as strings so we can convert them later.
	var productStateStr, productStatusStr string

	// Execute the query and scan the results.
	err := c.pool.QueryRow(ctx, query, req.GetId()).Scan(
		&product.Id,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category,
		&product.Tags,
		&createdAt,
		&updatedAt,
		&productStateStr,
		&productStatusStr,
		&variationData,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	// Convert timestamps to google.protobuf.Timestamp.
	product.CreatedAt = timestamppb.New(createdAt)
	product.UpdatedAt = timestamppb.New(updatedAt)

	// Convert the product state string to its enum.
	if stateVal, ok := pb.ProductState_value[productStateStr]; ok {
		product.ProductState = pb.ProductState(stateVal)
	} else {
		return nil, status.Errorf(codes.Internal, "invalid product state: %s", productStateStr)
	}

	// Convert the product status string to its enum.
	if statusVal, ok := pb.ProductStatus_value[productStatusStr]; ok {
		product.ProductStatus = pb.ProductStatus(statusVal)
	} else {
		return nil, status.Errorf(codes.Internal, "invalid product status: %s", productStatusStr)
	}

	// If variation data is present, decode it.
	if len(variationData) > 0 {
		// Log the original variation data for debugging.
		original := string(variationData)

		// Transform the JSON data to match the expected oneof field names in the proto.
		// Replace any capitalized keys ("Clothing", "Electronics", "Food") with the lowercase versions.
		transformed := original
		transformed = strings.Replace(transformed, `"Clothing":`, `"clothing":`, 1)
		transformed = strings.Replace(transformed, `"Electronics":`, `"electronics":`, 1)
		transformed = strings.Replace(transformed, `"Food":`, `"food":`, 1)

		// Unmarshal the transformed JSON into a temporary Product.
		var tmpProduct pb.Product
		unmarshaler := protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}
		if err := unmarshaler.Unmarshal([]byte(transformed), &tmpProduct); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to decode variation: %v", err)
		}

		// Assign the correct oneof field based on the type.
		switch v := tmpProduct.Variation.(type) {
		case *pb.Product_Clothing:
			product.Variation = &pb.Product_Clothing{Clothing: v.Clothing}
		case *pb.Product_Food:
			product.Variation = &pb.Product_Food{Food: v.Food}
		case *pb.Product_Electronics:
			product.Variation = &pb.Product_Electronics{Electronics: v.Electronics}
		default:
			// If the variation does not match any expected type, leave it as nil.
			product.Variation = nil
		}
	}

	// Return the product wrapped in a GetProductResponse.
	return &pb.GetProductResponse{
		Product: &product,
	}, nil
}

func (c *productController) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	// Define the SQL query to retrieve all products.
	query := `
		SELECT id, name, description, price, category, tags, created_at, updated_at, product_state, product_status, variation
		FROM products
	`

	// Execute the query.
	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}
	defer rows.Close()

	// Prepare a slice to store the resulting products.
	var products []*pb.Product

	// Loop through each row returned by the query.
	for rows.Next() {
		var product pb.Product
		var variationData []byte
		var createdAt, updatedAt time.Time

		// Scan the current row into our product variables.
		err := rows.Scan(
			&product.Id,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.Tags,
			&createdAt,
			&updatedAt,
			&product.ProductState,
			&product.ProductStatus,
			&variationData,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan product: %v", err)
		}

		// Convert the database timestamps to protobuf timestamps.
		product.CreatedAt = timestamppb.New(createdAt)
		product.UpdatedAt = timestamppb.New(updatedAt)

		// If variation data exists, unmarshal it using protojson.
		if len(variationData) > 0 {
			var tmpVariation pb.Product
			if err := protojson.Unmarshal(variationData, &tmpVariation); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to decode variation: %v", err)
			}
			// Depending on the oneof type, assign the appropriate variation field.
			switch v := tmpVariation.Variation.(type) {
			case *pb.Product_Clothing:
				product.Variation = &pb.Product_Clothing{Clothing: v.Clothing}
			case *pb.Product_Food:
				product.Variation = &pb.Product_Food{Food: v.Food}
			case *pb.Product_Electronics:
				product.Variation = &pb.Product_Electronics{Electronics: v.Electronics}
			default:
				product.Variation = nil
			}
		}

		// Add the product to the list.
		products = append(products, &product)
	}

	// Check for any errors encountered during row iteration.
	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error reading products: %v", err)
	}

	// Return the list of products in the response.
	return &pb.ListProductsResponse{
		Products: products,
	}, nil
}
