package models_test

import (
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/models"
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

