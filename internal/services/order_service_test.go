package services_test

import (
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/services"
	"github.com/Bimidu/ctse-order-service/internal/testutils"
	"github.com/google/uuid"
)

func TestOrderServiceCheckoutCreatesOrder(t *testing.T) {
	testutils.SetupTestDB(t)

	svc := services.NewOrderService()
	userID := "user-checkout"

	_, err := services.NewCartService().AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-9",
		Name:      "Checkout Item",
		Price:     12.5,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("expected add item to succeed, got error: %v", err)
	}

	order, err := svc.Checkout(userID)
	if err != nil {
		t.Fatalf("expected checkout to succeed, got error: %v", err)
	}
	if order.UserID != userID {
		t.Fatalf("expected user %s, got %s", userID, order.UserID)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 order item, got %d", len(order.Items))
	}
	if order.TotalAmount != 25.0 {
		t.Fatalf("expected total 25.0, got %.2f", order.TotalAmount)
	}
}

func TestOrderServiceCheckoutEmptyCartFails(t *testing.T) {
	testutils.SetupTestDB(t)

	svc := services.NewOrderService()
	if _, err := svc.Checkout("empty-user"); err == nil {
		t.Fatalf("expected checkout to fail for empty cart")
	}
}

func TestOrderServiceGetOrdersAndUpdateStatus(t *testing.T) {
	testutils.SetupTestDB(t)

	userID := "user-orders"
	order := createOrder(t, userID)

	svc := services.NewOrderService()
	orders, total, err := svc.GetUserOrders(userID, 10, 0)
	if err != nil {
		t.Fatalf("expected get user orders to succeed, got error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}

	updated, err := svc.UpdateStatus(order.ID.String(), models.StatusShipped)
	if err != nil {
		t.Fatalf("expected update status to succeed, got error: %v", err)
	}
	if updated.Status != models.StatusShipped {
		t.Fatalf("expected status shipped, got %s", updated.Status)
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

