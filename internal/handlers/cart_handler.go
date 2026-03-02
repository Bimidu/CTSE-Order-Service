package handlers

import (
	"net/http"

	"github.com/Bimidu/ctse-order-service/internal/middleware"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/Bimidu/ctse-order-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CartHandler struct {
	svc *services.CartService
}

func NewCartHandler() *CartHandler {
	return &CartHandler{svc: services.NewCartService()}
}

// AddItem godoc
// @Summary      Add item to cart
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body models.AddToCartRequest true "Cart item"
// @Success      201 {object} models.CartItem
// @Router       /api/v1/cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {
	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.svc.AddItem(middleware.GetUserID(c), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// GetCart godoc
// @Summary      Get current user's cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} models.CartResponse
// @Router       /api/v1/cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	cart, err := h.svc.GetCart(middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cart)
}

// UpdateItem godoc
// @Summary      Update cart item quantity
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path     string                     true "Cart item ID"
// @Param        body body     models.UpdateCartItemRequest true "New quantity"
// @Success      200  {object} models.CartItem
// @Router       /api/v1/cart/items/{id} [put]
func (h *CartHandler) UpdateItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.svc.UpdateItem(middleware.GetUserID(c), itemID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cart item not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

// RemoveItem godoc
// @Summary      Remove item from cart
// @Tags         cart
// @Security     BearerAuth
// @Param        id path string true "Cart item ID"
// @Success      204
// @Router       /api/v1/cart/items/{id} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	if err := h.svc.RemoveItem(middleware.GetUserID(c), itemID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "cart item not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ClearCart godoc
// @Summary      Clear all items from cart
// @Tags         cart
// @Security     BearerAuth
// @Success      204
// @Router       /api/v1/cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	if err := h.svc.ClearCart(middleware.GetUserID(c)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
