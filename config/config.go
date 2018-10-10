package config

import (
	"os"

	"github.com/spf13/viper"
)

func Init() {

	// Google
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("STEAM_GOOGLE_APPLICATION_CREDENTIALS"))
	}

	//
	viper.AutomaticEnv()
	viper.SetEnvPrefix("STEAM")

	// Rabbit
	viper.SetDefault("RABBIT_USER", "guest")
	viper.SetDefault("RABBIT_PASS", "guest")

	// Other
	viper.SetDefault("PORT", "8081")
	viper.SetDefault("ENV", "local")
	viper.SetDefault("MEMCACHE_DSN", "memcache:11211")
	viper.SetDefault("PATH", "/root")
	viper.SetDefault("MYSQL_DSN", "root@tcp(localhost:3306)/steam?parseTime=true")
	viper.SetDefault("DOMAIN", "https://gamedb.online")
	viper.SetDefault("SHORT_NAME", "GameDB")
}
