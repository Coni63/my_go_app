package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// User représente les informations utilisateur après authentification
type User struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
}

// AuthValidationResponse représente la réponse du service d'authentification
type AuthValidationResponse struct {
	IsValid bool   `json:"isValid"`
	User    User   `json:"user"`
	Message string `json:"message"`
}

// AuthRequest représente la requête envoyée au service d'authentification
type AuthRequest struct {
	Token string `json:"token"`
}

// AuthMiddleware est un middleware Gin pour l'authentification des requêtes
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupère le token de l'en-tête Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Access Denied: No token provided"})
			return
		}

		token := ""
		// Vérifie le format "Bearer TOKEN"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Access Denied: Invalid token format"})
			return
		}

		// URL du service d'authentification (à configurer via une variable d'environnement)
		authServiceURL := os.Getenv("AUTH_SERVICE_URL")
		if authServiceURL == "" {
			log.Println("AUTH_SERVICE_URL environment variable not set")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error: Authentication service URL not configured"})
			return
		}

		// Crée la requête HTTP pour le service d'authentification
		authReqBody, err := json.Marshal(AuthRequest{Token: token})
		if err != nil {
			log.Printf("Error marshalling auth request: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error while preparing authentication request"})
			return
		}

		resp, err := http.Post(authServiceURL, "application/json", bytes.NewBuffer(authReqBody))
		if err != nil {
			log.Printf("Error contacting auth service: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error contacting authentication service"})
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading auth service response: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			return
		}

		var authResponse AuthValidationResponse
		err = json.Unmarshal(bodyBytes, &authResponse)
		if err != nil {
			log.Printf("Error unmarshalling auth service response: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
			return
		}

		if authResponse.IsValid {
			// Stocke les informations utilisateur dans le contexte Gin pour les handlers suivants
			c.Set("user", authResponse.User)
			c.Next() // Passe au prochain handler (la logique de la route)
		} else {
			// Token invalide
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": authResponse.Message})
			return
		}
	}
}

// AuthorizeRolesMiddleware est un middleware Gin pour l'autorisation basée sur les rôles
func AuthorizeRolesMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAny, exists := c.Get("user")
		if !exists {
			// Cela ne devrait pas arriver si AuthMiddleware est appelé avant
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "User not authenticated"})
			return
		}
		user, ok := userAny.(User)
		if !ok {
			log.Printf("User context is not of type User: %v", userAny)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error: User context malformed"})
			return
		}

		hasPermission := false
		for _, allowedRole := range allowedRoles {
			for _, userRole := range user.Roles {
				if userRole == allowedRole {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Access Denied: Insufficient permissions"})
			return
		}
		c.Next() // Passe au prochain handler
	}
}
