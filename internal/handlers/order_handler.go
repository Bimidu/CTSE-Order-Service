package handlers

import (
	"net/http"
	"strconv"

	"github.com/Bimidu/ctse-order-service/internal/middleware"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc *services.OrderService
}

func NewOrderHandler() *OrderHandler {
	return &OrderHandler{svc: services.NewOrderService()}
}

// Checkout godoc
// @Summary      Checkout — create order from cart
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Success      201 {object} models.CheckoutResponse
// @Router       /api/v1/orders/checkout [post]
func (h *OrderHandler) Checkout(c *gin.Context) {
	order, err := h.svc.Checkout(middleware.GetUserID(c))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cart is empty" {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, models.CheckoutResponse{
		Order:   *order,
		Message: "Order placed successfully",
	})
}

// GetMyOrders godoc
// @Summary      Get current user's orders
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        limit  query int false "Page size (default 10)"
// @Param        offset query int false "Page offset (default 0)"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/orders [get]
func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit > 50 {
		limit = 50
	}

	orders, total, err := h.svc.GetUserOrders(middleware.GetUserID(c), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetOrder godoc
// @Summary      Get a specific order by ID
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Order ID"
// @Success      200 {object} models.Order
// @Router       /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	order, err := h.svc.GetOrder(middleware.GetUserID(c), c.Param("id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "order not found" || err.Error() == "invalid order id" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetAllOrders godoc (admin only)
// @Summary      Admin — get all orders
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        limit  query  int    false "Page size"
// @Param        offset query  int    false "Page offset"
// @Param        status query  string false "Filter by status"
// @Success      200 {object} map[string]interface{}
// @Router       /api/v1/admin/orders [get]
func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	statusFilter := c.Query("status")
	if limit > 100 {
		limit = 100
	}

	orders, total, err := h.svc.GetAllOrders(limit, offset, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateOrderStatus godoc (admin only)
// @Summary      Admin — update order status
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path string                         true "Order ID"
// @Param        body body models.UpdateOrderStatusRequest true "New status"
// @Success      200  {object} models.Order
// @Router       /api/v1/admin/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.svc.UpdateStatus(c.Param("id"), req.Status)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "order not found" || err.Error() == "invalid order id" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}
