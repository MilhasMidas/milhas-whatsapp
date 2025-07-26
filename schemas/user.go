package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(100);unique;not null"`
	Phone     string    `gorm:"type:varchar(20);not null"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"default:current_timestamp"`
	UpdatedAt time.Time `gorm:"default:current_timestamp"`
	// Relationships
	Subscriptions Subscription `gorm:"foreignKey:UserID"`
	WpNumbers     []WpNumber   `gorm:"foreignKey:UserID"`
}

type UserResponse struct {
	ID            uuid.UUID            `json:"id"`
	Name          string               `json:"name"`
	Email         string               `json:"email"`
	Phone         string               `json:"phone"`
	Password      string               `json:"password,omitempty"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdatedAt     time.Time            `json:"updatedAt"`
	Active        bool                 `json:"active"`
	Subscriptions SubscriptionResponse `json:"subscription,omitempty"`
	WpNumbers     []WpNumberResponse   `json:"wpNumbers,omitempty"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password,omitempty"`
}
