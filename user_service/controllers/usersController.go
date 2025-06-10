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

	middlewares "auth_service/middlewares"
	"auth_service/models"
	"auth_service/services"
)

type ApiController struct {
	DB    *gorm.DB
	Cache *ristretto.Cache
	Store *persistence.InMemoryStore
}

func NewApiController(db *gorm.DB, cache *ristretto.Cache, store *persistence.InMemoryStore) *ApiController {
	return &ApiController{
		DB:    db,
		Cache: cache,
		Store: store,
	}
}

// Example method
func (ac *ApiController) RegisterRoutes(r *gin.Engine) {
	cacheMiddleware := middlewares.NewCacheMiddleware(ac.Cache)

	r.GET("/user/api/v1/users", ac.GetAllUsers)
	r.GET("/user/api/v1/user/me", ac.GetMe)
	r.GET("/user/api/v1/user/:id", ac.GetUser)
	r.POST("/user/api/v1/user/", ac.CreateUser)
	r.GET("/user/api/v1/email/:email", ac.GetUserByEmail)
	r.PUT("/user/api/v1/user/:id", cacheMiddleware.Handler(), ac.PutUser)
	r.PATCH("/user/api/v1/user/:id", cacheMiddleware.Handler(), ac.PatchUser)
	r.DELETE("/user/api/v1/user/:id", cacheMiddleware.Handler(), ac.DeleteUser)
}

func (ac *ApiController) GetAllUsers(c *gin.Context) {
	// First, query the actual User models from the database
	users, err := services.GetAllUsers(ac.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Then convert to public view
	publicUsers := []models.UserPublicView{}
	for _, user := range users {
		publicUsers = append(publicUsers, user.ToPublicView())
	}

	c.JSON(http.StatusOK, publicUsers)
}

func (ac *ApiController) GetUser(c *gin.Context) {
	id := c.Param("id")

	// Check if it's a valid UUID
	_, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	// Query the actual User model
	user, err := services.GetUserById(ac.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Convert to public view
	publicView := user.ToPublicView()

	c.JSON(http.StatusOK, publicView)
}

func (ac *ApiController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")

	// Query the actual User model
	user, err := services.GetUserByEmail(ac.DB, email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Convert to public view
	publicView := user.ToPublicView()

	c.JSON(http.StatusOK, publicView)
}

func (ac *ApiController) PutUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)
	// Check if it's a valid UUID
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}
	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user"})
		return
	}

	// Get the request body
	var body struct {
		FirstName string `json:"firstname" binding:"required"`
		LastName  string `json:"lastname" binding:"required"`
		UserName  string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user data
	currentUser.FirstName = body.FirstName
	currentUser.LastName = body.LastName
	currentUser.UserName = body.UserName
	// Update other fields

	// Save the updated user
	if _, err := services.UpdateUser(ac.DB, currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Convert to private view
	privateView := currentUser.ToPrivateView()
	c.JSON(http.StatusNoContent, privateView)
}

func (ac *ApiController) PatchUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	// Check if it's a valid UUID
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this user"})
		return
	}

	// Get the request body
	var body struct {
		FirstName *string `json:"firstname"`
		LastName  *string `json:"lastname"`
		Username  *string `json:"username"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only the fields that were provided
	if body.FirstName != nil {
		currentUser.FirstName = *body.FirstName
	}
	if body.LastName != nil {
		currentUser.LastName = *body.LastName
	}
	if body.Username != nil {
		currentUser.UserName = *body.Username
	}

	// Save the updated user
	if _, err := services.UpdateUser(ac.DB, currentUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Convert to private view
	privateView := currentUser.ToPrivateView()

	c.JSON(http.StatusNoContent, privateView)
}

func (ac *ApiController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	// Check if it's a valid UUID
	userUuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if currentUser.ID != userUuid {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this user"})
		return
	}

	// Delete the user (GORM will set DeletedAt if the model uses gorm.DeletedAt)
	if err := services.DeleteUser(ac.DB, currentUser.ID.String()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{"message": "User deleted successfully"})
}

func (ac *ApiController) CreateUser(c *gin.Context) {
	var body struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindWith(&body, binding.Form); err != nil {
		// Check if it's a validation error
		if errs, ok := err.(validator.ValidationErrors); ok {
			// Create a more descriptive error message
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

	// Check if the user already exists
	if existingUser, _ := services.GetUserByEmail(ac.DB, body.Email); existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	user := models.User{Email: body.Email, Password: string(encryptedPassword)}
	createdUser, err := services.CreateUser(ac.DB, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, createdUser.ToPrivateView())
}

func (ac *ApiController) GetMe(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	c.JSON(http.StatusCreated, user.ToPrivateView())
}
