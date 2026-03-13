package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	ProductID string    `gorm:"not null" json:"product_id"`
	Name      string    `gorm:"not null" json:"name"`
	Price     float64   `gorm:"not null" json:"price"`
	Quantity  int       `gorm:"not null;check:quantity > 0" json:"quantity"`
	MoodTag   string    `json:"mood_tag,omitempty"`
	ImageURL  string    `json:"image_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddToCartRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	MoodTag   string  `json:"mood_tag"`
	ImageURL  string  `json:"image_url"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
	ItemCount  int        `json:"item_count"`
}
