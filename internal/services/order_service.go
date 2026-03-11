package services

import (
	"errors"
	"fmt"
	"log"

	"github.com/Bimidu/ctse-order-service/grpc/clients"
	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	productpb "github.com/Bimidu/ctse-order-service/proto/product"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct {
	cartSvc *CartService
}

func NewOrderService() *OrderService {
	return &OrderService{cartSvc: NewCartService()}
}

// Checkout converts the user's cart into a confirmed order.
// It validates stock via Product Service gRPC, creates the order,
// reduces stock, and notifies Auth Service to increment purchase count.
func (s *OrderService) Checkout(userID string) (*models.Order, error) {
	cart, err := s.cartSvc.GetCart(userID)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Validate stock via Product Service (if available)
	if clients.Product != nil {
		stockItems := make([]*productpb.StockItem, len(cart.Items))
		for i, item := range cart.Items {
			stockItems[i] = &productpb.StockItem{
				ProductId: item.ProductID,
				Quantity:  int32(item.Quantity),
			}
		}
		resp, err := clients.Product.ValidateStock(stockItems)
		if err != nil {
			log.Printf("Warning: stock validation failed: %v — proceeding without validation", err)
		} else if !resp.AllAvailable {
			unavailable := make([]string, len(resp.UnavailableItems))
			for i, u := range resp.UnavailableItems {
				unavailable[i] = fmt.Sprintf("%s (%s)", u.ProductId, u.Reason)
			}
			return nil, fmt.Errorf("some items are out of stock: %v", unavailable)
		}
	}

	// Build the order
	orderID := uuid.New()
	orderItems := make([]models.OrderItem, len(cart.Items))
	for i, item := range cart.Items {
		orderItems[i] = models.OrderItem{
			ID:        uuid.New(),
			OrderID:   orderID,
			ProductID: item.ProductID,
			Name:      item.Name,
			Price:     item.Price,
			Quantity:  item.Quantity,
			MoodTag:   item.MoodTag,
			ImageURL:  item.ImageURL,
		}
	}

	order := &models.Order{
		ID:          orderID,
		UserID:      userID,
		Status:      models.StatusPending,
		TotalAmount: cart.TotalPrice,
		Items:       orderItems,
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	if err := tx.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to clear cart after checkout: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fire-and-forget: reduce stock via Product Service
	if clients.Product != nil {
		stockItems := make([]*productpb.StockItem, len(cart.Items))
		for i, item := range cart.Items {
			stockItems[i] = &productpb.StockItem{ProductId: item.ProductID, Quantity: int32(item.Quantity)}
		}
		go func() {
			if _, err := clients.Product.ReduceStock(orderID.String(), stockItems); err != nil {
				log.Printf("Warning: failed to reduce stock for order %s: %v", orderID, err)
			}
		}()
	}

	// Fire-and-forget: increment purchase count in Auth Service
	if clients.Auth != nil {
		go func() {
			if _, err := clients.Auth.UpdatePurchaseCount(userID, 1); err != nil {
				log.Printf("Warning: failed to update purchase count for user %s: %v", userID, err)
			}
		}()
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID string, limit, offset int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	database.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)
	if err := database.DB.Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, total, nil
}

func (s *OrderService) GetOrder(userID, orderID string) (*models.Order, error) {
	id, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order id")
	}

	var order models.Order
	query := database.DB.Preload("Items").Where("id = ?", id)

	// Non-admin users can only see their own orders
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) GetAllOrders(limit, offset int, statusFilter string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := database.DB.Model(&models.Order{})
	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}
	query.Count(&total)

	if err := database.DB.Preload("Items").
		Where(func() *gorm.DB {
			if statusFilter != "" {
				return database.DB.Where("status = ?", statusFilter)
			}
			return database.DB
		}()).
		Order("created_at desc").
		Limit(limit).Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	return orders, total, nil
}

func (s *OrderService) UpdateStatus(orderID string, status models.OrderStatus) (*models.Order, error) {
	id, err := uuid.Parse(orderID)
	if err != nil {
		return nil, fmt.Errorf("invalid order id")
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("order not found")
		}
		return nil, err
	}

	order.Status = status
	if err := database.DB.Save(&order).Error; err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	if err := database.DB.Preload("Items").First(&order, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}
