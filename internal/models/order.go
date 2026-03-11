package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string      `gorm:"not null;index" json:"user_id"`
	Status      OrderStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TotalAmount float64     `gorm:"not null" json:"total_amount"`
	Items       []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID   uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID string    `gorm:"not null" json:"product_id"`
	Name      string    `gorm:"not null" json:"name"`
	Price     float64   `gorm:"not null" json:"price"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	MoodTag   string    `json:"mood_tag,omitempty"`
	ImageURL  string    `json:"image_url,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required,oneof=pending confirmed shipped delivered cancelled"`
}

type CheckoutResponse struct {
	Order   Order  `json:"order"`
	Message string `json:"message"`
}
