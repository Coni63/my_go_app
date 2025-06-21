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

type GroupController struct {
	DB    *gorm.DB
	Cache *ristretto.Cache
	Store *persistence.InMemoryStore
}

func NewGroupController(db *gorm.DB, cache *ristretto.Cache, store *persistence.InMemoryStore) *GroupController {
	return &GroupController{
		DB:    db,
		Cache: cache,
		Store: store,
	}
}

func (gc *GroupController) RegisterRoutes(r *gin.Engine) {
	r.GET("/user/api/v1/groups", gc.GetAllGroups)
	r.GET("/user/api/v1/group/:id", gc.GetGroup)
	r.POST("/user/api/v1/group", gc.CreateGroup)
	r.PUT("/user/api/v1/group/:id", gc.UpdateGroup)
	r.DELETE("/user/api/v1/group/:id", gc.DeleteGroup)
}

func (gc *GroupController) GetAllGroups(c *gin.Context) {
	groups, err := services.GetAllGroups(gc.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

func (gc *GroupController) GetGroup(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	group, err := services.GetGroupById(gc.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (gc *GroupController) CreateGroup(c *gin.Context) {
	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdGroup, err := services.CreateGroup(gc.DB, &group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}
	c.JSON(http.StatusCreated, createdGroup)
}

func (gc *GroupController) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var group models.Group
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedGroup, err := services.UpdateGroup(gc.DB, id, &group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}
	c.JSON(http.StatusOK, updatedGroup)
}

func (gc *GroupController) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := services.DeleteGroup(gc.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
