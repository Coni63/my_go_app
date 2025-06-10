package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"auth_service/config"
	"auth_service/models"
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

// AuthValidationRequest est la structure de la requête JSON entrante
type AuthValidationRequest struct {
	Token string `json:"token"`
}

// AuthValidationResponse est la structure de la réponse JSON sortante
type AuthValidationResponse struct {
	IsValid bool   `json:"isValid"`
	UserId  string `json:"userId,omitempty"` // Optionnel, peut être rempli si le token est valide
	Message string `json:"message,omitempty"`
}

// CheckTokenValidity est la fonction que tu as fournie, légèrement adaptée pour ne pas aborter le contexte Gin
// mais plutôt retourner les erreurs.
func CheckTokenValidity(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// TOKEN_SECRET_KEY doit être la même que celle utilisée pour signer le token
		secretKey := os.Getenv("TOKEN_SECRET_KEY")
		if secretKey == "" {
			return nil, fmt.Errorf("TOKEN_SECRET_KEY environment variable not set")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		// Loguer l'erreur pour le débogage mais ne pas exposer directement au client
		fmt.Printf("Token parsing error: %v\n", err)
		return nil, fmt.Errorf("invalid or expired token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims format")
	}

	// Expiry check - jwt.Parse avec la v5 gère normalement déjà cela.
	// Cependant, si tu veux une vérification explicite pour un message plus spécifique :
	if expFloat, ok := claims["exp"].(float64); ok {
		if float64(time.Now().Unix()) > expFloat {
			return nil, fmt.Errorf("token expired")
		}
	} else {
		// Optionnel: si "exp" n'est pas présent ou n'est pas un float64, on peut considérer cela comme une erreur
		// ou laisser passer si les tokens peuvent être éternels (déconseillé).
		// Pour cet exemple, on ne force pas le "exp" pour la validité si ce n'est pas déjà géré par Parse.
	}

	return claims, nil
}

func (ac *ApiController) RegisterRoutes(r *gin.Engine) {
	r.POST("/auth/api/v1/signup", ac.Signup)
	r.POST("/auth/api/v1/login", ac.Login)
	r.POST("/auth/api/v1/reset-password", ac.ResetPassword)
	r.POST("/auth/api/v1/validate-token", ac.ValidateToken)
}

func (ac *ApiController) Signup(c *gin.Context) {
	// URL du service d'authentification (à configurer via une variable d'environnement)
	authServiceURL := os.Getenv("USER_SERVICE_URL")
	if authServiceURL == "" {
		log.Println("USER_SERVICE_URL environment variable not set")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error: User service URL not configured"})
		return
	}

	url := fmt.Sprintf("%s/user/api/v1/user/", authServiceURL)
	resp, err := http.Post(url, "application/json", c.Request.Body)
	if err != nil {
		log.Printf("Error contacting auth service: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error contacting authentication service"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading user service response: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error reading response from user service"})
		return
	}

	// Si le user service a échoué, on renvoie son message d’erreur
	if resp.StatusCode >= 400 {
		log.Printf("User service returned error: %s", string(respBody))
		c.AbortWithStatusJSON(resp.StatusCode, gin.H{"message": string(respBody)})
		return
	}

	// Sinon, on renvoie le contenu directement
	c.Data(resp.StatusCode, "application/json", respBody)
}

func (ac *ApiController) Login(c *gin.Context) {
	var body struct {
		Email    string `form:"email" binding:"required,email"`
		Password string `form:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		var errors []string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
			}
		} else {
			errors = append(errors, err.Error())
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Call user service to retrieve user by email
	userServiceURL := os.Getenv("USER_SERVICE_URL")
	if userServiceURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User service not configured"})
		return
	}
	url := fmt.Sprintf("%s/user/api/v1/email/%s", userServiceURL, body.Email)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"}) // avoid leaking which failed
		return
	}
	defer resp.Body.Close()

	// TODO: Create custom modul with email, pwd and uuid

	// First, find the user by name only
	var existingUser models.User
	if err := ac.DB.First(&existingUser, "email = ?", body.Email).Error; err != nil {
		// User not found
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Now verify the password separately using bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(body.Password)); err != nil {
		// Password doesn't match
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       existingUser.ID.String(),
		"exp":       time.Now().Add(config.TokenTTL).Unix(), // Token expires in 24 hours
		"iat":       time.Now().Unix(),
		"iss":       os.Getenv("TOKEN_ISSUER"),
		"username":  existingUser.UserName,
		"firstname": existingUser.FirstName,
		"lastname":  existingUser.LastName,
		"email":     existingUser.Email,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("TOKEN_SECRET_KEY")))
	if err != nil {
		log.Println("Error signing token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// If we get here, user is authenticated
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (ac *ApiController) ResetPassword(c *gin.Context) {
	// TODO: Implement password reset functionality
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented"})
}

func (ac *ApiController) ValidateToken(c *gin.Context) {
	// ValidateTokenHandler est le handler Gin pour l'endpoint /api/auth/validate-token
	var req AuthValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthValidationResponse{
			IsValid: false,
			Message: "Bad request: token field missing or malformed",
		})
		return
	}

	claims, err := CheckTokenValidity(req.Token)
	if err != nil {
		// Si CheckTokenValidity retourne une erreur, le token est invalide ou expiré
		c.JSON(http.StatusUnauthorized, AuthValidationResponse{
			IsValid: false,
			UserId:  "",          // Pas d'ID utilisateur car le token est invalide
			Message: err.Error(), // Renvoie le message d'erreur spécifique de CheckTokenValidity
		})
		return
	}

	// Si nous arrivons ici, le token est valide.

	// Load from DB
	userID, ok := claims["sub"].(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token -- missing sub claim"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, AuthValidationResponse{
		IsValid: true,
		UserId:  userID,
		Message: "Token is valid",
	})
}
