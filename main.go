package main

import (
	"os"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"gorm.io/gorm"

	"my_app/auth_service/controllers"
	"my_app/shared_modules/packages/initializers"
	"my_app/shared_modules/packages/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Store *persistence.InMemoryStore
var Cache *ristretto.Cache
var DB *gorm.DB

func init() {
	config := initializers.LoadConfig("config.yaml")
	initializers.LoadEnvVariables()
	initializers.ConnectToDB(DB, os.Getenv("DSN"))
	initializers.InitCache(Store, Cache, config.CacheTTL, config.CacheSize, config.MaxCost, config.BufferItems)
}

func main() {
	r := gin.Default()

	r.Use(middlewares.PrometheusStatusCodeMiddleware())
	r.GET("/auth/api/v1/metrics", gin.WrapH(promhttp.Handler()))

	apiController := controllers.NewApiController(DB, Cache, Store)
	apiController.RegisterRoutes(r)

	r.Run(os.Getenv("ADDR"))
}
