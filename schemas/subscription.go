package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	gorm.Model
	Id           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Product      string    `gorm:"type:varchar(100);not null"`
	PurchaseDate time.Time `gorm:"type:timestamp;not null"`
	ExpiryDate   time.Time `gorm:"type:timestamp;not null"`
	Active       bool      `gorm:"default:true"`
	CreatedAt    time.Time `gorm:"default:current_timestamp"`
	UpdatedAt    time.Time `gorm:"default:current_timestamp"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
}

type SubscriptionResponse struct {
	ID           uuid.UUID `json:"id"`
	Product      string    `json:"product"`
	PurchaseDate time.Time `json:"purchaseDate"`
	ExpiryDate   time.Time `json:"expiryDate"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
