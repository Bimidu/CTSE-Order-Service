package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/Bimidu/ctse-order-service/internal/config"
	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	orderpb "github.com/Bimidu/ctse-order-service/proto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrderGRPCServer struct {
	orderpb.UnimplementedOrderServiceServer
}

func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	id, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order id: %v", err)
	}

	var order models.Order
	if err := database.DB.Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	return toProtoOrder(&order), nil
}

func (s *OrderGRPCServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	var orders []models.Order
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	offset := int(req.Offset)

	var total int64
	database.DB.Model(&models.Order{}).Where("user_id = ?", req.UserId).Count(&total)
	database.DB.Preload("Items").Where("user_id = ?", req.UserId).
		Order("created_at desc").Limit(limit).Offset(offset).Find(&orders)

	protoOrders := make([]*orderpb.OrderResponse, len(orders))
	for i, o := range orders {
		protoOrders[i] = toProtoOrder(&o)
	}

	return &orderpb.GetUserOrdersResponse{
		Orders: protoOrders,
		Total:  int32(total),
	}, nil
}

func (s *OrderGRPCServer) GetUserOrderCount(ctx context.Context, req *orderpb.GetUserOrderCountRequest) (*orderpb.GetUserOrderCountResponse, error) {
	var count int64
	database.DB.Model(&models.Order{}).Where("user_id = ?", req.UserId).Count(&count)
	return &orderpb.GetUserOrderCountResponse{Count: int32(count)}, nil
}

func toProtoOrder(o *models.Order) *orderpb.OrderResponse {
	items := make([]*orderpb.OrderItem, len(o.Items))
	for i, item := range o.Items {
		items[i] = &orderpb.OrderItem{
			ProductId: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  int32(item.Quantity),
			MoodTag:   item.MoodTag,
		}
	}
	return &orderpb.OrderResponse{
		Id:          o.ID.String(),
		UserId:      o.UserID,
		Status:      string(o.Status),
		TotalAmount: o.TotalAmount,
		Items:       items,
		CreatedAt:   o.CreatedAt.String(),
	}
}

func Start() {
	port := config.App.GRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", port, err)
	}

	s := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(s, &OrderGRPCServer{})

	log.Printf("Order gRPC server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
