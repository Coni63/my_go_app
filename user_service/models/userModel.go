package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	FirstName   string    `json:"firstname"`
	LastName    string    `json:"lastname"`
	UserName    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string
	Groups      []UserGroup      `gorm:"foreignKey:UserID"`
	Permissions []UserPermission `gorm:"foreignKey:UserID"`
}

type UserPublicView struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

type UserPrivateView struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"firstname"`
	LastName    string    `json:"lastname"`
	UserName    string    `json:"username"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Groups      []string  `json:"groups"`
	Permissions []string  `json:"permissions"`
}

func (u *User) ToPublicView() UserPublicView {
	return UserPublicView{
		ID:       u.ID.String(),
		UserName: u.UserName,
	}
}

func (u *User) ToPrivateView() UserPrivateView {
	// Get group names
	groupNames := make([]string, len(u.Groups))
	for i, group := range u.Groups {
		groupNames[i] = group.Group.Name
	}

	// Get permission names
	permissionNames := make([]string, len(u.Permissions))
	for i, perm := range u.Permissions {
		permissionNames[i] = perm.Permission.Name
	}

	return UserPrivateView{
		ID:          u.ID.String(),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		UserName:    u.UserName,
		Email:       u.Email,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Groups:      groupNames,
		Permissions: permissionNames,
	}
}
