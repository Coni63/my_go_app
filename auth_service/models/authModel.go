package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthUser struct {
	gorm.Model
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email     string    `gorm:"uniqueIndex"`
	Password  string
	LastLogin time.Time      `gorm:"default:NULL"` // Optional field to track last login time
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Soft delete support
}

type RetrievePasswordModel struct {
	gorm.Model
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Email      string    `gorm:"uniqueIndex"`
	ResetToken string    `gorm:"uniqueIndex"`
	ExpiresAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Used       bool      `gorm:"default:false"` // Indicate if the reset token has been used
}

type AuthAuditLog struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Activity  string    `gorm:"type:text"`        // Description of the activity like login / logout / password changed / password request
	IPAddress string    `gorm:"type:varchar(45)"` // To store IPv4 and IPv6 addresses
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
