package server_test

import (
	"context"
	"testing"

	"github.com/Bimidu/ctse-order-service/grpc/server"
	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/testutils"
	orderpb "github.com/Bimidu/ctse-order-service/proto/order"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOrderGRPCServerGetOrder(t *testing.T) {
	testutils.SetupTestDB(t)

	order := createOrder(t, "grpc-user")
	srv := &server.OrderGRPCServer{}

	resp, err := srv.GetOrder(context.Background(), &orderpb.GetOrderRequest{OrderId: order.ID.String()})
	if err != nil {
		t.Fatalf("expected gRPC get order to succeed, got error: %v", err)
	}
	if resp.UserId != order.UserID {
		t.Fatalf("expected user %s, got %s", order.UserID, resp.UserId)
	}
}

func TestOrderGRPCServerGetOrderInvalid(t *testing.T) {
	testutils.SetupTestDB(t)

	srv := &server.OrderGRPCServer{}
	_, err := srv.GetOrder(context.Background(), &orderpb.GetOrderRequest{OrderId: "bad-id"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}

func TestOrderGRPCServerGetUserOrders(t *testing.T) {
	testutils.SetupTestDB(t)

	createOrder(t, "user-a")
	createOrder(t, "user-a")
	createOrder(t, "user-b")

	srv := &server.OrderGRPCServer{}
	resp, err := srv.GetUserOrders(context.Background(), &orderpb.GetUserOrdersRequest{UserId: "user-a", Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("expected get user orders to succeed, got error: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(resp.Orders))
	}
}

func createOrder(t *testing.T, userID string) models.Order {
	t.Helper()

	order := models.Order{
		ID:          uuid.New(),
		UserID:      userID,
		Status:      models.StatusPending,
		TotalAmount: 15.0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	item := models.OrderItem{
		ID:        uuid.New(),
		OrderID:   order.ID,
		ProductID: "prod-x",
		Name:      "Seed Item",
		Price:     15.0,
		Quantity:  1,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("failed to create order item: %v", err)
	}

	order.Items = []models.OrderItem{item}
	return order
}

