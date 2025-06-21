package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"unique;not null"`
	Description string
}

type UserPermission struct {
	gorm.Model
	UserID       uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey"`
	User         User       `gorm:"foreignKey:UserID"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
}

type GroupPermission struct {
	gorm.Model
	GroupID      uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Group        Group      `gorm:"foreignKey:GroupID"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
}
