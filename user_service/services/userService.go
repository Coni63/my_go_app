package services

import (
	"errors"
	"fmt"
	"my_app/user_service/models"

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
	result := db.Preload("Groups.Group").
		Preload("Permissions.Permission").
		First(&user, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

func GetUserByEmail(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	result := db.Preload("Groups.Group").
		Preload("Permissions.Permission").
		First(&user, "email = ?", email)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

func GetUserPermissions(db *gorm.DB, userId string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := db.Model(&models.UserPermission{}).
		Joins("JOIN permissions ON user_permissions.permission_id = permissions.id").
		Where("user_permissions.user_id = ?", userId).
		Find(&permissions).Error
	return permissions, err
}

func GetUserGroupPermissions(db *gorm.DB, userId string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := db.Model(&models.UserGroup{}).
		Joins("JOIN group_permissions ON user_groups.group_id = group_permissions.group_id").
		Joins("JOIN permissions ON group_permissions.permission_id = permissions.id").
		Where("user_groups.user_id = ?", userId).
		Find(&permissions).Error
	return permissions, err
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

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with id %s", id)
	}

	return nil
}

// Group Service Methods
func GetAllGroups(db *gorm.DB) ([]models.Group, error) {
	groups := []models.Group{}
	err := db.Find(&groups).Error
	return groups, err
}

func GetGroupById(db *gorm.DB, id string) (*models.Group, error) {
	var group models.Group
	result := db.First(&group, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("group not found")
		}
		return nil, result.Error
	}

	return &group, nil
}

func CreateGroup(db *gorm.DB, group *models.Group) (*models.Group, error) {
	if err := db.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func UpdateGroup(db *gorm.DB, id string, group *models.Group) (*models.Group, error) {
	var existingGroup models.Group
	if err := db.First(&existingGroup, "id = ?", id).Error; err != nil {
		return nil, err
	}

	existingGroup.Name = group.Name
	if err := db.Save(&existingGroup).Error; err != nil {
		return nil, err
	}

	return &existingGroup, nil
}

func DeleteGroup(db *gorm.DB, id string) error {
	result := db.Delete(&models.Group{}, id)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no group found with id %s", id)
	}

	return nil
}

// User-Group Relationship Methods
func AddUserToGroup(db *gorm.DB, userGroup *models.UserGroup) error {
	if err := db.Create(userGroup).Error; err != nil {
		return err
	}
	return nil
}

func RemoveUserFromGroup(db *gorm.DB, userId string, groupId string) error {
	result := db.Where("user_id = ? AND group_id = ?", userId, groupId).Delete(&models.UserGroup{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no user-group relationship found")
	}

	return nil
}

func GetUserGroups(db *gorm.DB, userId string) ([]models.Group, error) {
	var groups []models.Group
	err := db.Model(&models.UserGroup{}).
		Joins("JOIN groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userId).
		Find(&groups).Error
	return groups, err
}

func GetGroupUsers(db *gorm.DB, groupId string) ([]models.User, error) {
	var users []models.User
	err := db.Model(&models.UserGroup{}).
		Joins("JOIN users ON user_groups.user_id = users.id").
		Where("user_groups.group_id = ?", groupId).
		Find(&users).Error
	return users, err
}
