package main

import (
	"auth_service/config"
	"auth_service/controllers"
	"os"

	shared_middlewares "github.com/coni63/my_app/shared_modules/packages/middlewares"

	"github.com/coni63/my_app/shared_modules/packages/initializers"
	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Store *persistence.InMemoryStore
var Cache *ristretto.Cache
var DB *gorm.DB

func init() {

	initializers.LoadEnvVariables()
	initializers.ConnectToDB(DB, os.Getenv("DSN"))
	initializers.InitCache(Store, Cache, config.CacheTTL, config.CacheSize, config.MaxCost, config.BufferItems)
}

func main() {
	r := gin.Default()

	r.Use(shared_middlewares.PrometheusStatusCodeMiddleware())
	r.GET("/user/api/v1/metrics", gin.WrapH(promhttp.Handler()))

	apiController := controllers.NewApiController(DB, Cache, Store)
	apiController.RegisterRoutes(r)

	r.Run(os.Getenv("ADDR"))
}
