package main

import (
	"os"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
	"gorm.io/gorm"

	"my_app/shared_modules/packages/initializers"
	"my_app/shared_modules/packages/middlewares"
	group_controllers "my_app/user_service/controllers"
	user_controllers "my_app/user_service/controllers"

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
	initializers.InitCache(Store, Cache, config.Cache.CacheTTL, config.Cache.CacheSize, config.Cache.MaxCost, config.Cache.BufferItems)
}

func main() {
	r := gin.Default()

	r.Use(middlewares.PrometheusStatusCodeMiddleware())
	r.GET("/auth/api/v1/metrics", gin.WrapH(promhttp.Handler()))

	// Initialize user controller
	userController := user_controllers.NewUserController(DB, Cache, Store)
	userController.RegisterRoutes(r)

	groupController := group_controllers.NewGroupController(DB, Cache, Store)
	groupController.RegisterRoutes(r)

	// Initialize user group controller
	userGroupController := user_controllers.NewUserGroupController(DB, Cache, Store)
	userGroupController.RegisterRoutes(r)

	r.Run(os.Getenv("ADDR"))
}
