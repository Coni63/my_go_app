package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserGroup struct {
	gorm.Model
	ID      uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	UserID  uuid.UUID `gorm:"type:uuid;column:user_id"`
	GroupID uuid.UUID `gorm:"type:uuid;column:group_id"`
	User    User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Group   Group     `gorm:"foreignKey:GroupID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
