package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WpNumber struct {
	gorm.Model
	Id        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Number    string    `gorm:"type:varchar(20);not null;unique"`
	Server    string    `gorm:"type:varchar(100);not null"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"default:current_timestamp"`
	UpdatedAt time.Time `gorm:"default:current_timestamp"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	// Relationships
	Webhooks []Webhook `gorm:"foreignKey:WpNumberID"`
}

type WpNumberResponse struct {
	ID        uuid.UUID `json:"id"`
	Number    string    `json:"number"`
	Server    string    `json:"server"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	UserID    uuid.UUID `json:"userId"`
	Webhooks  []Webhook `json:"webhooks"`
}
