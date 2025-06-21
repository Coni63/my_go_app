package interfaces

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	duration, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
	Cache struct {
		Enabled     bool          `yaml:"enabled"`
		CacheSize   int64         `yaml:"cache_size"`
		MaxCost     int64         `yaml:"max_cost"`
		BufferItems int64         `yaml:"buffer_items"`
		CacheTTL    time.Duration `yaml:"cache_ttl"`
	} `yaml:"cache"`
}
