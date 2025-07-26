package schemas

import (
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"
)

type Webhook struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Events     []string  `gorm:"type:text[]"`
	CreatedAt  time.Time `gorm:"default:current_timestamp"`
	UpdatedAt  time.Time `gorm:"default:current_timestamp"`
	Active     bool      `gorm:"default:true"`
	URL        string    `gorm:"type:varchar(255);not null"`
	WpNumberID uuid.UUID `gorm:"type:uuid;not null"`
}

type WebhookResponse struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Active     bool      `json:"active"`
	URL        string    `json:"url"`
	WpNumberID uuid.UUID `json:"wpNumberId"`
}
