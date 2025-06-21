package controllers

import (
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"my_app/user_service/models"
	"my_app/user_service/services"
)

type UserGroupController struct {
	DB    *gorm.DB
	Cache *ristretto.Cache
	Store *persistence.InMemoryStore
}

func NewUserGroupController(db *gorm.DB, cache *ristretto.Cache, store *persistence.InMemoryStore) *UserGroupController {
	return &UserGroupController{
		DB:    db,
		Cache: cache,
		Store: store,
	}
}

func (ugc *UserGroupController) RegisterRoutes(r *gin.Engine) {
	r.POST("/user/api/v1/user/:userId/group/:groupId", ugc.AddUserToGroup)
	r.DELETE("/user/api/v1/user/:userId/group/:groupId", ugc.RemoveUserFromGroup)
	r.GET("/user/api/v1/user/:userId/groups", ugc.GetUserGroups)
	r.GET("/user/api/v1/group/:groupId/users", ugc.GetGroupUsers)
}

func (ugc *UserGroupController) AddUserToGroup(c *gin.Context) {
	userId := c.Param("userId")
	groupId := c.Param("groupId")

	if _, err := uuid.Parse(userId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	if _, err := uuid.Parse(groupId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group UUID format"})
		return
	}

	userUUID, _ := uuid.Parse(userId)
	groupUUID, _ := uuid.Parse(groupId)
	userGroup := models.UserGroup{
		UserID:  userUUID,
		GroupID: groupUUID,
	}

	if err := services.AddUserToGroup(ugc.DB, &userGroup); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to group"})
		return
	}

	c.JSON(http.StatusCreated, userGroup)
}

func (ugc *UserGroupController) RemoveUserFromGroup(c *gin.Context) {
	userId := c.Param("userId")
	groupId := c.Param("groupId")

	if _, err := uuid.Parse(userId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	if _, err := uuid.Parse(groupId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group UUID format"})
		return
	}

	if err := services.RemoveUserFromGroup(ugc.DB, userId, groupId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user from group"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ugc *UserGroupController) GetUserGroups(c *gin.Context) {
	userId := c.Param("userId")

	if _, err := uuid.Parse(userId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user UUID format"})
		return
	}

	groups, err := services.GetUserGroups(ugc.DB, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user groups"})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (ugc *UserGroupController) GetGroupUsers(c *gin.Context) {
	groupId := c.Param("groupId")

	if _, err := uuid.Parse(groupId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group UUID format"})
		return
	}

	users, err := services.GetGroupUsers(ugc.DB, groupId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
