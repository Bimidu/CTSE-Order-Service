package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/handlers"
	"github.com/Bimidu/ctse-order-service/internal/middleware"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/testutils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCartHandlerAddItemAndCheckout(t *testing.T) {
	testutils.SetupTestDB(t)
	gin.SetMode(gin.TestMode)

	r := setupTestRouter("user-42", "user")

	addBody := `{"product_id":"prod-1","name":"Test","price":5.5,"quantity":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewBufferString(addBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.Code)
	}

	checkoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/orders/checkout", nil)
	checkoutResp := httptest.NewRecorder()
	r.ServeHTTP(checkoutResp, checkoutReq)
	if checkoutResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", checkoutResp.Code)
	}
}

func TestAdminUpdateOrderStatusHandler(t *testing.T) {
	testutils.SetupTestDB(t)
	gin.SetMode(gin.TestMode)

	order := createOrder(t, "admin-user")
	r := setupTestRouter("admin-user", "admin")

	payload, _ := json.Marshal(models.UpdateOrderStatusRequest{Status: models.StatusDelivered})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/orders/"+order.ID.String()+"/status", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}
}

func setupTestRouter(userID, role string) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	})

	cartH := handlers.NewCartHandler()
	orderH := handlers.NewOrderHandler()

	v1 := r.Group("/api/v1")
	{
		cart := v1.Group("/cart")
		{
			cart.POST("/items", cartH.AddItem)
			cart.GET("", cartH.GetCart)
			cart.PUT("/items/:id", cartH.UpdateItem)
			cart.DELETE("/items/:id", cartH.RemoveItem)
			cart.DELETE("", cartH.ClearCart)
		}

		orders := v1.Group("/orders")
		{
			orders.POST("/checkout", orderH.Checkout)
			orders.GET("", orderH.GetMyOrders)
			orders.GET("/:id", orderH.GetOrder)
		}

		admin := v1.Group("/admin")
		admin.Use(middleware.RoleRequired("admin"))
		{
			admin.GET("/orders", orderH.GetAllOrders)
			admin.PUT("/orders/:id/status", orderH.UpdateOrderStatus)
		}
	}

	return r
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

