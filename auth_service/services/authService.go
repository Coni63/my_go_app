package services

import (
	"auth_service/models"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CreateUser(db *gorm.DB, user *models.User) (*models.User, error) {
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserById(db *gorm.DB, id string) (*models.User, error) {
	var user models.User
	result := db.First(&user, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error // Return other errors normally
	}

	return &user, nil // Return found user
}

func GetUserByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	result := db.First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error // Return other errors normally
	}

	return &user, nil // Return found user
}

func GetAllUsers(db *gorm.DB) ([]models.User, error) {
	users := []models.User{}
	err := db.Find(&users).Error
	return users, err
}

func UpdateUser(db *gorm.DB, user *models.User) (*models.User, error) {
	if err := db.Save(&user).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

func DeleteUser(db *gorm.DB, id string) error {
	result := db.Delete(&models.User{}, id)

	// Check if the delete was successful
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with id %s", id)
	}

	return nil
}
