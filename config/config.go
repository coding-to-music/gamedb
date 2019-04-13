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

	Config.MySQLHost.Set("MYSQL_HOST")
	Config.MySQLPort.Set("MYSQL_PORT")
	Config.MySQLUsername.Set("MYSQL_USERNAME")
	Config.MySQLPassword.Set("MYSQL_PASSWORD")
	Config.MySQLDatabase.Set("MYSQL_DATABASE")

	Config.RecaptchaPublic = os.Getenv(prefix + "RECAPTCHA_PUBLIC")
	Config.RecaptchaPrivate = os.Getenv(prefix + "RECAPTCHA_PRIVATE")

	Config.TwitchClientID = os.Getenv(prefix + "TWITCH_CLIENT_ID")
	Config.TwitchClientSecret = os.Getenv(prefix + "TWITCH_CLIENT_SECRET")

	Config.InfluxURL = os.Getenv(prefix + "INFLUX_URL")
	Config.InfluxUsername = os.Getenv(prefix + "INFLUX_USERNAME")
	Config.InfluxPassword = os.Getenv(prefix + "INFLUX_PASSWORD")

	Config.MongoHost = os.Getenv(prefix + "MONGO_HOST")
	Config.MongoPort = os.Getenv(prefix + "MONGO_PORT")
	Config.MongoUsername = os.Getenv(prefix + "MONGO_USERNAME")
	Config.MongoPassword = os.Getenv(prefix + "MONGO_PASSWORD")
	Config.MongoDatabase = os.Getenv(prefix + "MONGO_DATABASE")

	Config.TwitterAccessToken = os.Getenv(prefix + "TWITTER_ACCESS_TOKEN")
	Config.TwitterAccessTokenSecret = os.Getenv(prefix + "TWITTER_ACCESS_TOKEN_SECRET")
	Config.TwitterConsumerKey = os.Getenv(prefix + "TWITTER_CONSUMER_KEY")
	Config.TwitterConsumerSecret = os.Getenv(prefix + "TWITTER_CONSUMER_SECRET")

	Config.PatreonSecret = os.Getenv(prefix + "PATREON_WEBOOK_SECRET")
	Config.PatreonClientID = os.Getenv(prefix + "PATREON_CLIENT_ID")
	Config.PatreonClientSecret = os.Getenv(prefix + "PATREON_CLIENT_SECRET")

	Config.DiscordClientID = os.Getenv(prefix + "DISCORD_CLIENT_ID")
	Config.DiscordSescret = os.Getenv(prefix + "DISCORD_SECRET")
	Config.DiscordBotToken = os.Getenv(prefix + "DISCORD_BOT_TOKEN")

	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken = os.Getenv(prefix + "GITHUB_TOKEN")
	Config.GoogleBucket = os.Getenv(prefix + "GOOGLE_BUCKET")
	Config.GoogleProject = os.Getenv(prefix + "GOOGLE_PROJECT")
	Config.GameDBDirectory.Set("PATH")
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
	Config.GameDBDirectory.SetDefault("/root")
	Config.NewReleaseDays = 14

	switch Config.Environment.Get() {
	case EnvProd:

		Config.MemcacheDSN.SetDefault("memcache:11211")
		// Config.EnableConsumers.SetDefault("0")

	case EnvLocal:

		Config.MemcacheDSN.SetDefault("localhost:11211")
		Config.PatreonSecret = "EZTRjtID_1LUmgnQ4_WWuWIQbfj4QA1JtqYMq4prcq_kDvNdEXlgj2K7JyLwNXfd"

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

	DiscordClientID string
	DiscordSescret  string
	DiscordBotToken string

	InfluxURL      string
	InfluxPassword string
	InfluxUsername string

	InstagramPassword ConfigItem
	InstagramUsername ConfigItem

	MongoHost     string
	MongoPort     string
	MongoUsername string
	MongoPassword string
	MongoDatabase string

	MySQLHost     ConfigItem
	MySQLPort     ConfigItem
	MySQLUsername ConfigItem
	MySQLPassword ConfigItem
	MySQLDatabase ConfigItem

	RabbitUsername      ConfigItem
	RabbitPassword      ConfigItem
	RabbitHost          string
	RabbitPort          string
	RabbitManagmentPort string

	RecaptchaPrivate string
	RecaptchaPublic  string

	SessionAuthentication string
	SessionEncryption     string

	TwitchClientID     string
	TwitchClientSecret string

	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	TwitterConsumerKey       string
	TwitterConsumerSecret    string

	PatreonSecret       string
	PatreonClientID     string
	PatreonClientSecret string

	Environment     ConfigItem
	GameDBDirectory ConfigItem
	GameDBDomain    ConfigItem
	GameDBShortName ConfigItem
	GithubToken     string
	GoogleBucket    string
	GoogleProject   string
	MemcacheDSN     ConfigItem
	SendGridAPIKey  string
	SteamAPIKey     string
	WebserverPort   ConfigItem
	EnableWebserver ConfigItem
	EnableConsumers ConfigItem
	CommitHash      string
	NewReleaseDays  int
}

func (c BaseConfig) RabbitDSN() string {
	return "amqp://" + c.RabbitUsername.Get() + ":" + c.RabbitPassword.Get() + "@" + c.RabbitHost + ":" + c.RabbitPort
}

func (c BaseConfig) MySQLDNS() string {
	return c.MySQLUsername.Get() + ":" + c.MySQLPassword.Get() + "@tcp(" + c.MySQLHost.Get() + ":" + c.MySQLPort.Get() + ")/" + c.MySQLDatabase.Get()
}

func (c BaseConfig) MongoDSN() string {
	return "mongodb://" + c.MongoHost + ":" + c.MongoPort
}

func (c BaseConfig) RabbitAPI(values url.Values) string {
	return "http://" + c.RabbitHost + ":" + c.RabbitManagmentPort + "/api/overview?" + values.Encode()
}

func (c BaseConfig) ListenOn() string {
	return "0.0.0.0:" + c.WebserverPort.Get()
}

func (c BaseConfig) HasMemcache() bool {
	return c.MemcacheDSN.Get() != ""
}

func (c BaseConfig) IsLocal() bool {
	return c.Environment.Get() == EnvLocal
}

func (c BaseConfig) IsProd() bool {
	return c.Environment.Get() == EnvProd
}

func (c BaseConfig) IsConsumer() bool {
	return c.Environment.Get() == EnvConsumer
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
	b, err := strconv.ParseBool(ci.Get())
	if err != nil {
		fmt.Println(err)
	}
	return b
}

func (ci ConfigItem) GetInt() int {
	i, err := strconv.Atoi(ci.Get())
	if err != nil {
		fmt.Println(err)
	}
	return i
}
