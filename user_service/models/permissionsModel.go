package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permissions struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"permissionName"`
	Description string    `json:"permissionDescription"`
}
