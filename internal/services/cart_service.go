package services

import (
	"errors"
	"fmt"

	"github.com/Bimidu/ctse-order-service/internal/database"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartService struct{}

func NewCartService() *CartService {
	return &CartService{}
}

func (s *CartService) AddItem(userID string, req *models.AddToCartRequest) (*models.CartItem, error) {
	var existing models.CartItem
	err := database.DB.Where("user_id = ? AND product_id = ?", userID, req.ProductID).First(&existing).Error

	if err == nil {
		// Item already exists — increment quantity
		existing.Quantity += req.Quantity
		if err := database.DB.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
		return &existing, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	item := &models.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: req.ProductID,
		Name:      req.Name,
		Price:     req.Price,
		Quantity:  req.Quantity,
		MoodTag:   req.MoodTag,
		ImageURL:  req.ImageURL,
	}

	if err := database.DB.Create(item).Error; err != nil {
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}
	return item, nil
}

func (s *CartService) GetCart(userID string) (*models.CartResponse, error) {
	var items []models.CartItem
	if err := database.DB.Where("user_id = ?", userID).Order("created_at asc").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	var totalPrice float64
	for _, item := range items {
		totalPrice += item.Price * float64(item.Quantity)
	}

	return &models.CartResponse{
		Items:      items,
		TotalPrice: totalPrice,
		ItemCount:  len(items),
	}, nil
}

func (s *CartService) UpdateItem(userID string, itemID uuid.UUID, req *models.UpdateCartItemRequest) (*models.CartItem, error) {
	var item models.CartItem
	if err := database.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cart item not found")
		}
		return nil, err
	}

	item.Quantity = req.Quantity
	if err := database.DB.Save(&item).Error; err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return &item, nil
}

func (s *CartService) RemoveItem(userID string, itemID uuid.UUID) error {
	result := database.DB.Where("id = ? AND user_id = ?", itemID, userID).Delete(&models.CartItem{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove cart item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("cart item not found")
	}
	return nil
}

func (s *CartService) ClearCart(userID string) error {
	if err := database.DB.Where("user_id = ?", userID).Delete(&models.CartItem{}).Error; err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}
