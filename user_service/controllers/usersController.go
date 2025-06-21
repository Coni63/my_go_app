package controllers

import (
	"fmt"
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"my_app/user_service/models"
	"my_app/user_service/services"
)

type UserController struct {
	DB    *gorm.DB
	Cache *ristretto.Cache
	Store *persistence.InMemoryStore
}

func NewUserController(db *gorm.DB, cache *ristretto.Cache, store *persistence.InMemoryStore) *UserController {
	return &UserController{
		DB:    db,
		Cache: cache,
		Store: store,
	}
}

func (uc *UserController) RegisterRoutes(r *gin.Engine) {
	r.GET("/user/api/v1/my-profile", uc.GetMe)
	r.GET("/user/api/v1/users", uc.GetAllUsers)
	r.GET("/user/api/v1/user/:id", uc.GetUser)
	r.POST("/user/api/v1/user/", uc.CreateUser)
	r.PUT("/user/api/v1/user/:id", uc.PutUser)
	r.PATCH("/user/api/v1/user/:id", uc.PatchUser)
	r.DELETE("/user/api/v1/user/:id", uc.DeleteUser)
	r.GET("/user/api/v1/user/:id/permissions", uc.GetUserPermissions)
	r.GET("/user/api/v1/user/:id/group-permissions", uc.GetUserGroupPermissions)
}

func (uc *UserController) GetAllUsers(c *gin.Context) {
	users, err := services.GetAllUsers(uc.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	publicUsers := []models.UserPublicView{}
	for _, user := range users {
		publicUsers = append(publicUsers, user.ToPublicView())
	}

	c.JSON(http.StatusOK, publicUsers)
}

func (uc *UserController) GetUser(c *gin.Context) {
	id := c.Param("id")
	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	user, err := services.GetUserById(uc.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToPublicView())
}

func (uc *UserController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user, err := services.GetUserByEmail(uc.DB, email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToPublicView())
}

func (uc *UserController) PutUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}
	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user"})
		return
	}

	var body struct {
		FirstName string `json:"firstname" binding:"required"`
		LastName  string `json:"lastname" binding:"required"`
		UserName  string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser.FirstName = body.FirstName
	currentUser.LastName = body.LastName
	currentUser.UserName = body.UserName

	if _, err := services.UpdateUser(uc.DB, currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusNoContent, currentUser.ToPrivateView())
}

func (uc *UserController) PatchUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user"})
		return
	}

	var body struct {
		FirstName *string `json:"firstname"`
		LastName  *string `json:"lastname"`
		Username  *string `json:"username"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.FirstName != nil {
		currentUser.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		currentUser.LastName = *body.LastName
	}
	if body.Username != nil {
		currentUser.UserName = *body.Username
	}

	if _, err := services.UpdateUser(uc.DB, currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusNoContent, currentUser.ToPrivateView())
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this user"})
		return
	}

	if err := services.DeleteUser(uc.DB, currentUser.ID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted successfully"})
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var body struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindWith(&body, binding.Form); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, e := range errs {
				errorMessages = append(errorMessages, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}
		return
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if existingUser, _ := services.GetUserByEmail(uc.DB, body.Email); existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	user := models.User{Email: body.Email, Password: string(encryptedPassword)}
	createdUser, err := services.CreateUser(uc.DB, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, createdUser.ToPrivateView())
}

func (uc *UserController) GetMe(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	userWithPerms, err := services.GetUserById(uc.DB, user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
		return
	}
	c.JSON(http.StatusOK, userWithPerms.ToPrivateView())
}

func (uc *UserController) GetUserPermissions(c *gin.Context) {
	id := c.Param("id")
	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	permissions, err := services.GetUserPermissions(uc.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (uc *UserController) GetUserGroupPermissions(c *gin.Context) {
	id := c.Param("id")
	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	permissions, err := services.GetUserGroupPermissions(uc.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group permissions"})
		return
	}

	c.JSON(http.StatusOK, permissions)
}
