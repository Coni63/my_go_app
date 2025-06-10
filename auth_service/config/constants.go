package config

import (
	"time"
)

var (
	TokenTTL          = 15 * time.Minute // Token expiration time
	CacheTTL          = 5 * time.Minute  // Cache TTL for user tokens
	CacheSize   int64 = 1e7              // Maximum count of items of the cache
	MaxCost     int64 = 1 << 28          // Max memory usage in bytes (256MB)
	BufferItems int64 = 64               // Number of items to buffer before writing to the cache
)
