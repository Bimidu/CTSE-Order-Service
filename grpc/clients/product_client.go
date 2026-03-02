package clients

import (
	"context"
	"log"
	"time"

	"github.com/Bimidu/ctse-order-service/internal/config"
	productpb "github.com/Bimidu/ctse-order-service/proto/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	conn   *grpc.ClientConn
	client productpb.ProductServiceClient
}

var Product *ProductClient

func InitProductClient() {
	addr := config.App.ProductServiceAddr
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: could not connect to product service at %s: %v", addr, err)
		return
	}
	Product = &ProductClient{
		conn:   conn,
		client: productpb.NewProductServiceClient(conn),
	}
	log.Printf("Product gRPC client connected to %s", addr)
}

func (p *ProductClient) GetProduct(productID string) (*productpb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.client.GetProduct(ctx, &productpb.GetProductRequest{ProductId: productID})
}

func (p *ProductClient) ValidateStock(items []*productpb.StockItem) (*productpb.ValidateStockResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return p.client.ValidateStock(ctx, &productpb.ValidateStockRequest{Items: items})
}

func (p *ProductClient) ReduceStock(orderID string, items []*productpb.StockItem) (*productpb.ReduceStockResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return p.client.ReduceStock(ctx, &productpb.ReduceStockRequest{
		OrderId: orderID,
		Items:   items,
	})
}

func (p *ProductClient) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}
