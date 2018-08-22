package config

import (
	"github.com/spf13/viper"
)

func Init() {

	viper.AutomaticEnv()
	viper.SetEnvPrefix("STEAM")

	// Rabbit
	viper.SetDefault("RABBIT_USER", "guest")
	viper.SetDefault("RABBIT_PASS", "guest")

	// Other
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("ENV", "local")
	viper.SetDefault("MEMCACHE_DSN", "memcache:11211")
	viper.SetDefault("STEAM_PATH", "/root")
	viper.SetDefault("STEAM_MYSQL_DSN", "root@tcp(localhost:3306)/steam?parseTime=true")
}
