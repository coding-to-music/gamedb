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
	Config.AdminUser = os.Getenv(prefix + "ADMIN_USER")
	Config.AdminPass = os.Getenv(prefix + "ADMIN_PASS")
	Config.AdminName = os.Getenv(prefix + "ADMIN_NAME")
	Config.AdminEmail = os.Getenv(prefix + "ADMIN_EMAIL")

	Config.RabbitUser.Set("RABBIT_USER")
	Config.RabbitPass.Set("RABBIT_PASS")
	Config.RabbitHost = os.Getenv(prefix + "RABBIT_HOST")
	Config.RabbitPort = os.Getenv(prefix + "RABBIT_PORT")
	Config.RabbitManPort = os.Getenv(prefix + "RABBIT_MANAGEMENT_PORT")

	Config.SessionAuthentication = os.Getenv(prefix + "SESSION_AUTHENTICATION")
	Config.SessionEncryption = os.Getenv(prefix + "SESSION_ENCRYPTION")

	Config.InstagramUsername.Set("INSTAGRAM_USERNAME")
	Config.InstagramPassword.Set("INSTAGRAM_PASSWORD")

	Config.RecaptchaPublic = os.Getenv(prefix + "RECAPTCHA_PUBLIC")
	Config.RecaptchaPrivate = os.Getenv(prefix + "RECAPTCHA_PRIVATE")

	Config.DiscordBotToken = os.Getenv(prefix + "DISCORD_BOT_TOKEN")
	Config.Domain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken = os.Getenv(prefix + "GITHUB_TOKEN")
	Config.GoogleAppCreds = os.Getenv(prefix + "GOOGLE_APPLICATION_CREDENTIALS")
	Config.GoogleBucket = os.Getenv(prefix + "GOOGLE_BUCKET")
	Config.GoogleProject = os.Getenv(prefix + "GOOGLE_PROJECT")
	Config.MemcacheDSN.Set("MEMCACHE_DSN")
	Config.MySQLDSN.Set("MYSQL_DSN")
	Config.Path.Set("PATH")
	Config.RollbarPrivateKey = os.Getenv(prefix + "ROLLBAR_PRIVATE")
	Config.SendGridAPIKey = os.Getenv(prefix + "SENDGRID")
	Config.ShortName.Set("SHORT_NAME")
	Config.SteamAPIKey = os.Getenv(prefix + "API_KEY")
	Config.WebServerPort.Set("PORT")

	// Fallbacks
	Config.ShortName.SetFallback("GameDB")
	Config.InstagramUsername.SetFallback("gamedb.online")

	if Config.IsLocal() {

		Config.RabbitUser.SetFallback("guest")
		Config.RabbitPass.SetFallback("guest")

		Config.WebServerPort.SetFallback("8081")
		Config.MemcacheDSN.SetFallback("memcache:11211")
		Config.MySQLDSN.SetFallback("root@tcp(localhost:3306)/steam")
		Config.Domain.SetFallback("http://localhost:8081")

	} else if Config.IsProd() {

		Config.Path.SetFallback("/root")
		Config.Domain.SetFallback("https://gamedb.online")

	} else {
		fmt.Println("Missing env")
		os.Exit(1)
	}
}

type BaseConfig struct {
	RabbitUser            ConfigItem
	RabbitPass            ConfigItem
	RabbitHost            string
	RabbitPort            string
	RabbitManPort         string
	WebServerPort         ConfigItem
	ShortName             ConfigItem
	GoogleProject         string
	MySQLDSN              ConfigItem
	MemcacheDSN           ConfigItem
	SteamAPIKey           string
	GoogleBucket          string
	Environment           ConfigItem
	RollbarPrivateKey     string
	SessionAuthentication string
	SessionEncryption     string
	InstagramUsername     ConfigItem
	InstagramPassword     ConfigItem
	AdminUser             string
	AdminPass             string
	AdminName             string
	AdminEmail            string
	DiscordBotToken       string
	GithubToken           string
	RecaptchaPublic       string
	RecaptchaPrivate      string
	SendGridAPIKey        string
	GoogleAppCreds        string
	Path                  ConfigItem
	Domain                ConfigItem
}

func (c BaseConfig) RabbitDSN() string {
	return "amqp://" + c.RabbitUser.Get() + ":" + c.RabbitPass.Get() + "@" + c.RabbitHost + ":" + c.RabbitPort
}

func (c BaseConfig) RabbitAPI(values url.Values) string {
	return "http://" + c.RabbitHost + ":" + c.RabbitManPort + "/api/overview?" + values.Encode()
}

func (c BaseConfig) ListenOn() string {
	return "0.0.0.0:" + c.WebServerPort.Get()
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
