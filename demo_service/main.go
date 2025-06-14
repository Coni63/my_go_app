package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Get the value from the environment variable
	message := os.Getenv("MY_MESSAGE")
	if message == "" {
		message = "Default Hello from ENV" // fallback value
	}

	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
			"env":     message,
		})
	})

	// Run the server on port 8080
	r.Run(":8080")
}
