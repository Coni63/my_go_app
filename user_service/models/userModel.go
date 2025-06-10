package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string
}

type UserPublicView struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
}

type UserPrivateView struct {
	ID        string    `json:"id"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	UserName  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// No password field
}

// Serializers for User model
// These methods convert the User model to its public and private views.
func (u *User) ToPublicView() UserPublicView {
	return UserPublicView{
		ID:       u.ID.String(),
		UserName: u.UserName,
	}
}

func (u *User) ToPrivateView() UserPrivateView {
	return UserPrivateView{
		ID:        u.ID.String(),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserName:  u.UserName,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
