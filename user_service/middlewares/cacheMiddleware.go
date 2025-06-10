package middlewares

import (
	"log"
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-gonic/gin"
)

type CacheMiddleware struct {
	Cache *ristretto.Cache
}

func NewCacheMiddleware(cache *ristretto.Cache) *CacheMiddleware {
	return &CacheMiddleware{Cache: cache}
}

func getTokenFromHeader(c *gin.Context) string {
	tokenString := c.Request.Header.Get("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
		c.Abort()
		return ""
	}

	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	return tokenString
}

func (cm *CacheMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Vérifie la méthode HTTP après l'exécution du handler
		// pour s'assurer que l'opération de modification a eu lieu.
		// On invalide le cache seulement si la requête a réussi (statut 2xx).
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			switch c.Request.Method {
			case http.MethodPut, http.MethodPatch, http.MethodDelete:
				token := getTokenFromHeader(c) // Récupère le token
				if token != "" {
					cm.Cache.Del(token)
				} else {
					log.Println("Warning: No token found in header for cache invalidation.")
				}
			}
		}
	}
}
