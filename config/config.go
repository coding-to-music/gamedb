package config

import (
	"fmt"
	"net/url"
	"os"
)

const EnvProd = "production"
const EnvLocal = "local"

const prefix = "STEAM_"

var Config BaseConfig

func init() {

	// Set configs from environment variables
	Config.AdminUsername = os.Getenv(prefix + "ADMIN_USER")
	Config.AdminPassword = os.Getenv(prefix + "ADMIN_PASS")
	Config.AdminName = os.Getenv(prefix + "ADMIN_NAME")
	Config.AdminEmail = os.Getenv(prefix + "ADMIN_EMAIL")

	Config.RabbitUsername.Set("RABBIT_USER")
	Config.RabbitPassword.Set("RABBIT_PASS")
	Config.RabbitHost = os.Getenv(prefix + "RABBIT_HOST")
	Config.RabbitPort = os.Getenv(prefix + "RABBIT_PORT")
	Config.RabbitManagmentPort = os.Getenv(prefix + "RABBIT_MANAGEMENT_PORT")

	Config.SessionAuthentication = os.Getenv(prefix + "SESSION_AUTHENTICATION")
	Config.SessionEncryption = os.Getenv(prefix + "SESSION_ENCRYPTION")

	Config.InstagramUsername.Set("INSTAGRAM_USERNAME")
	Config.InstagramPassword.Set("INSTAGRAM_PASSWORD")

	Config.RecaptchaPublic = os.Getenv(prefix + "RECAPTCHA_PUBLIC")
	Config.RecaptchaPrivate = os.Getenv(prefix + "RECAPTCHA_PRIVATE")

	Config.DiscordBotToken = os.Getenv(prefix + "DISCORD_BOT_TOKEN")
	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken = os.Getenv(prefix + "GITHUB_TOKEN")
	Config.GoogleBucket = os.Getenv(prefix + "GOOGLE_BUCKET")
	Config.GoogleProject = os.Getenv(prefix + "GOOGLE_PROJECT")
	Config.MemcacheDSN.Set("MEMCACHE_DSN")
	Config.MySQLDSN.Set("MYSQL_DSN")
	Config.GameDBDirectory.Set("PATH")
	Config.RollbarPrivateKey = os.Getenv(prefix + "ROLLBAR_PRIVATE")
	Config.SendGridAPIKey = os.Getenv(prefix + "SENDGRID")
	Config.GameDBShortName.Set("SHORT_NAME")
	Config.SteamAPIKey = os.Getenv(prefix + "API_KEY")
	Config.WebserverPort.Set("PORT")

	// Fallbacks
	Config.GameDBShortName.SetFallback("GameDB")
	Config.InstagramUsername.SetFallback("gamedb.online")
	Config.WebserverPort.SetFallback("8081")

	if Config.IsLocal() {

		Config.RabbitUsername.SetFallback("guest")
		Config.RabbitPassword.SetFallback("guest")

		Config.MemcacheDSN.SetFallback("memcache:11211")
		Config.MySQLDSN.SetFallback("root@tcp(localhost:3306)/steam")
		Config.GameDBDomain.SetFallback("http://localhost:8081")

	} else if Config.IsProd() {

		Config.GameDBDirectory.SetFallback("/root")
		Config.GameDBDomain.SetFallback("https://gamedb.online")

	} else {
		fmt.Println("Missing env")
		os.Exit(1)
	}
}

type BaseConfig struct {
	AdminEmail    string
	AdminName     string
	AdminPassword string
	AdminUsername string

	RabbitUsername      ConfigItem
	RabbitPassword      ConfigItem
	RabbitHost          string
	RabbitPort          string
	RabbitManagmentPort string

	SessionAuthentication string
	SessionEncryption     string

	InstagramPassword ConfigItem
	InstagramUsername ConfigItem

	RecaptchaPrivate string
	RecaptchaPublic  string

	DiscordBotToken   string
	Environment       ConfigItem
	GameDBDirectory   ConfigItem
	GameDBDomain      ConfigItem
	GameDBShortName   ConfigItem
	GithubToken       string
	GoogleBucket      string
	GoogleProject     string
	MemcacheDSN       ConfigItem
	MySQLDSN          ConfigItem
	RollbarPrivateKey string
	SendGridAPIKey    string
	SteamAPIKey       string
	WebserverPort     ConfigItem
}

func (c BaseConfig) RabbitDSN() string {
	return "amqp://" + c.RabbitUsername.Get() + ":" + c.RabbitPassword.Get() + "@" + c.RabbitHost + ":" + c.RabbitPort
}

func (c BaseConfig) RabbitAPI(values url.Values) string {
	return "http://" + c.RabbitHost + ":" + c.RabbitManagmentPort + "/api/overview?" + values.Encode()
}

func (c BaseConfig) ListenOn() string {
	return "0.0.0.0:" + c.WebserverPort.Get()
}

func (c BaseConfig) IsLocal() bool {
	return c.Environment.Get() == EnvLocal
}

func (c BaseConfig) IsProd() bool {
	return c.Environment.Get() == EnvProd
}

// ConfigItem
type ConfigItem struct {
	Value    string
	Fallback string
}

func (ci *ConfigItem) Set(environment string) {
	env := os.Getenv(prefix + environment)
	if env != "" {
		ci.Value = env
	} else {
		ci.Value = os.Getenv(environment)
	}
}

func (ci *ConfigItem) SetFallback(fallback string) {
	ci.Fallback = fallback
}

func (ci ConfigItem) Get() string {
	if ci.Value != "" {
		return ci.Value
	}
	return ci.Fallback
}
