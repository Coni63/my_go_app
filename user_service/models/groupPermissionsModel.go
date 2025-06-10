package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupPermissions struct {
	gorm.Model
	ID           uuid.UUID   `gorm:"type:uuid;default:gen_random_uuid()"`
	GroupID      uuid.UUID   `json:"groupId"`
	PermissionID uuid.UUID   `json:"permissionId"`
	Group        Group       `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Permission   Permissions `gorm:"foreignKey:PermissionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
