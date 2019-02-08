package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const EnvProd = "production"
const EnvLocal = "local"
const EnvConsumer = "consumer"

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

	Config.TwitterAccessToken = os.Getenv(prefix + "TWITTER_ACCESS_TOKEN")
	Config.TwitterAccessTokenSecret = os.Getenv(prefix + "TWITTER_ACCESS_TOKEN_SECRET")
	Config.TwitterConsumerKey = os.Getenv(prefix + "TWITTER_CONSUMER_KEY")
	Config.TwitterConsumerSecret = os.Getenv(prefix + "TWITTER_CONSUMER_SECRET")

	Config.DiscordBotToken = os.Getenv(prefix + "DISCORD_BOT_TOKEN")
	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken = os.Getenv(prefix + "GITHUB_TOKEN")
	Config.GoogleBucket = os.Getenv(prefix + "GOOGLE_BUCKET")
	Config.GoogleProject = os.Getenv(prefix + "GOOGLE_PROJECT")
	Config.MySQLDSN.Set("MYSQL_DSN")
	Config.GameDBDirectory.Set("PATH")
	Config.RollbarPrivateKey = os.Getenv(prefix + "ROLLBAR_PRIVATE")
	Config.SendGridAPIKey = os.Getenv(prefix + "SENDGRID")
	Config.GameDBShortName.Set("SHORT_NAME")
	Config.SteamAPIKey = os.Getenv(prefix + "API_KEY")
	Config.WebserverPort.Set("PORT")

	// Defaults
	Config.GameDBShortName.SetDefault("GameDB")
	Config.InstagramUsername.SetDefault("gamedb.online")
	Config.WebserverPort.SetDefault("8081")
	Config.EnableWebserver.SetDefault("1")
	Config.EnableConsumers.SetDefault("1")
	Config.GameDBDomain.SetDefault("https://gamedb.online")
	Config.MemcacheDSN.SetDefault("memcache:11211")
	Config.GameDBDirectory.SetDefault("/root")

	switch Config.Environment.Get() {
	case EnvProd:

	case EnvLocal:

		Config.RabbitUsername.SetDefault("guest")
		Config.RabbitPassword.SetDefault("guest")
		Config.MemcacheDSN.SetDefault("localhost:11211")
		Config.MySQLDSN.SetDefault("root@tcp(localhost:3306)/steam")
		Config.GameDBDomain.SetDefault("http://localhost:8081")

	case EnvConsumer:

		Config.EnableWebserver.SetDefault("0")

	default:
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

	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	TwitterConsumerKey       string
	TwitterConsumerSecret    string

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
	EnableWebserver   ConfigItem
	EnableConsumers   ConfigItem
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
	value        string
	defaultValue string
}

func (ci *ConfigItem) Set(environment string) {
	environment = strings.TrimPrefix(environment, prefix)
	env := os.Getenv(prefix + environment)
	if env != "" {
		ci.value = env
	} else {
		ci.value = os.Getenv(environment)
	}
}

func (ci *ConfigItem) SetDefault(defaultValue string) {
	ci.defaultValue = defaultValue
}

func (ci ConfigItem) Get() string {
	if ci.value != "" {
		return ci.value
	}
	return ci.defaultValue
}

func (ci ConfigItem) GetBool() bool {
	b, _ := strconv.ParseBool(ci.Get())
	return b
}
