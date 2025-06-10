package initializers

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cache/persistence"
)

func InitCache(Store *persistence.InMemoryStore, Cache *ristretto.Cache, CacheTTL time.Duration, CacheSize int64, MaxCost int64, BufferItems int64) {
	var err error

	Store = persistence.NewInMemoryStore(CacheTTL)

	Cache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: CacheSize,
		MaxCost:     MaxCost,
		BufferItems: BufferItems,
	})
	if err != nil {
		panic(err)
	}
}
