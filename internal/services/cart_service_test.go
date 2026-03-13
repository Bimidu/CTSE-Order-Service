package services_test

import (
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/services"
	"github.com/Bimidu/ctse-order-service/internal/testutils"
	"github.com/google/uuid"
)

func TestAddToCartRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.AddToCartRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.AddToCartRequest{
				ProductID: "prod-001",
				Name:      "Test Product",
				Price:     9.99,
				Quantity:  2,
			},
			wantErr: false,
		},
		{
			name: "zero price should fail binding",
			req: models.AddToCartRequest{
				ProductID: "prod-001",
				Name:      "Test Product",
				Price:     0,
				Quantity:  1,
			},
			wantErr: true,
		},
		{
			name: "zero quantity should fail binding",
			req: models.AddToCartRequest{
				ProductID: "prod-001",
				Name:      "Test Product",
				Price:     9.99,
				Quantity:  0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.req.Price <= 0 || tt.req.Quantity < 1 ||
				tt.req.ProductID == "" || tt.req.Name == ""
			if hasError != tt.wantErr {
				t.Errorf("expected error=%v, got error=%v", tt.wantErr, hasError)
			}
		})
	}
}

func TestOrderStatusConstants(t *testing.T) {
	statuses := []models.OrderStatus{
		models.StatusPending,
		models.StatusConfirmed,
		models.StatusShipped,
		models.StatusDelivered,
		models.StatusCancelled,
	}

	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("order status should not be empty")
		}
	}
}

func TestCartServiceAddItemAndIncrement(t *testing.T) {
	testutils.SetupTestDB(t)

	svc := services.NewCartService()
	userID := "user-1"

	item, err := svc.AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-1",
		Name:      "Test Product",
		Price:     10.0,
		Quantity:  2,
	})
	if err != nil {
		t.Fatalf("expected add item to succeed, got error: %v", err)
	}
	if item.Quantity != 2 {
		t.Fatalf("expected quantity 2, got %d", item.Quantity)
	}

	item, err = svc.AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-1",
		Name:      "Test Product",
		Price:     10.0,
		Quantity:  3,
	})
	if err != nil {
		t.Fatalf("expected add item (increment) to succeed, got error: %v", err)
	}
	if item.Quantity != 5 {
		t.Fatalf("expected quantity 5, got %d", item.Quantity)
	}
}

func TestCartServiceUpdateAndRemoveItem(t *testing.T) {
	testutils.SetupTestDB(t)

	svc := services.NewCartService()
	userID := "user-1"

	item, err := svc.AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-2",
		Name:      "Another Product",
		Price:     7.5,
		Quantity:  1,
	})
	if err != nil {
		t.Fatalf("expected add item to succeed, got error: %v", err)
	}

	updated, err := svc.UpdateItem(userID, item.ID, &models.UpdateCartItemRequest{Quantity: 4})
	if err != nil {
		t.Fatalf("expected update to succeed, got error: %v", err)
	}
	if updated.Quantity != 4 {
		t.Fatalf("expected quantity 4, got %d", updated.Quantity)
	}

	err = svc.RemoveItem(userID, item.ID)
	if err != nil {
		t.Fatalf("expected remove to succeed, got error: %v", err)
	}

	missingID := uuid.New()
	if err := svc.RemoveItem(userID, missingID); err == nil {
		t.Fatalf("expected remove to fail for missing item")
	}
}

func TestCartServiceGetCartTotals(t *testing.T) {
	testutils.SetupTestDB(t)

	svc := services.NewCartService()
	userID := "user-2"

	_, _ = svc.AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-a",
		Name:      "Item A",
		Price:     5.0,
		Quantity:  2,
	})
	_, _ = svc.AddItem(userID, &models.AddToCartRequest{
		ProductID: "prod-b",
		Name:      "Item B",
		Price:     3.0,
		Quantity:  1,
	})

	cart, err := svc.GetCart(userID)
	if err != nil {
		t.Fatalf("expected get cart to succeed, got error: %v", err)
	}
	if cart.ItemCount != 2 {
		t.Fatalf("expected 2 items, got %d", cart.ItemCount)
	}
	if cart.TotalPrice != 13.0 {
		t.Fatalf("expected total 13.0, got %.2f", cart.TotalPrice)
	}
}


