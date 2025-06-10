package services

import (
	"auth_service/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateAuthUser(db *gorm.DB, email string, hashed_password string) (*models.AuthUser, error) {
	user := &models.AuthUser{
		Email:    email,
		Password: hashed_password,
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		if err := EmitAuditLog(tx, user.ID, "User Registered", ""); err != nil {
			return fmt.Errorf("failed to emit audit log: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil

}

func GetAuthUserById(db *gorm.DB, id string) (*models.AuthUser, error) {
	var user models.AuthUser
	result := db.First(&user, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error // Return other errors normally
	}

	return &user, nil // Return found user
}

func GetAuthUserByEmail(db *gorm.DB, email string) (*models.AuthUser, error) {
	var user models.AuthUser
	result := db.First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error // Return other errors normally
	}
	return &user, nil
}

func UpdateAuthUserLastLogin(db *gorm.DB, email string) (*models.AuthUser, error) {
	var updatedUser *models.AuthUser

	err := db.Transaction(func(tx *gorm.DB) error {
		// Load user within the transaction
		user, err := GetAuthUserByEmail(tx, email)
		if err != nil {
			return err // rollback
		}

		user.LastLogin = time.Now()

		// Save within the same transaction
		if err := tx.Save(&user).Error; err != nil {
			return err // rollback
		}

		updatedUser = user

		if err := EmitAuditLog(tx, user.ID, "User Logged In", ""); err != nil {
			return fmt.Errorf("failed to emit audit log: %w", err)
		}

		return nil // commit
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func UpdateAuthUserPasswordUpdate(db *gorm.DB, email string, hashedPassword string) (*models.AuthUser, error) {
	var updatedUser *models.AuthUser

	err := db.Transaction(func(tx *gorm.DB) error {
		user, err := GetAuthUserByEmail(tx, email)
		if err != nil {
			return err // rollback
		}

		user.Password = hashedPassword

		if err := tx.Save(&user).Error; err != nil {
			return err // rollback
		}

		updatedUser = user

		if err := EmitAuditLog(tx, user.ID, "Password Updated", ""); err != nil {
			return fmt.Errorf("failed to emit audit log: %w", err)
		}

		return nil // commit
	})

	if err != nil {
		return nil, err
	}
	return updatedUser, nil
}

func DeleteUser(db *gorm.DB, email string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		user, err := GetAuthUserByEmail(tx, email)
		if err != nil {
			return err // rollback
		}

		user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}

		if err := tx.Save(&user).Error; err != nil {
			return err // rollback
		}

		if err := EmitAuditLog(tx, user.ID, "User Deleted", ""); err != nil {
			return fmt.Errorf("failed to emit audit log: %w", err)
		}

		return nil // commit
	})
}

func EmitAuditLog(db *gorm.DB, userID uuid.UUID, activity string, ipAddress string) error {
	log := &models.AuthAuditLog{
		UserID:    userID,
		Activity:  activity,
		IPAddress: ipAddress,
	}

	if err := db.Create(log).Error; err != nil {
		return err
	}
	return nil
}
