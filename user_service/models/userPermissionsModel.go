package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserPermissions struct {
	gorm.Model
	ID           uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid()"`
	UserID       uuid.UUID   `gorm:"type:uuid;column:user_id"` // Changed to match Go naming convention
	PermissionID uuid.UUID   `gorm:"type:uuid;column:permission_id"`
	User         User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Permission   Permissions `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
