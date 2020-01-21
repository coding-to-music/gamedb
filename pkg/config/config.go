package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const EnvProd = "production"
const EnvLocal = "local"
const EnvConsumer = "consumer"

const prefix = "STEAM_"

var Config BaseConfig

type BaseConfig struct {
	// Admin
	AdminName    ConfigItem
	AdminEmail   ConfigItem
	AdminSteamID ConfigItem

	// Discord
	DiscordClientID      ConfigItem
	DiscordClientSescret ConfigItem

	DiscordRelayBotToken   ConfigItem
	DiscordChatBotToken    ConfigItem
	DiscordChangesBotToken ConfigItem

	// Facebook
	FacebookAppID     ConfigItem
	FacebookAppSecret ConfigItem

	// Google
	GoogleBucket  ConfigItem
	GoogleProject ConfigItem

	GoogleOauthClientID     ConfigItem
	GoogleOauthClientSecret ConfigItem

	// Influx
	InfluxURL      ConfigItem
	InfluxPassword ConfigItem
	InfluxUsername ConfigItem

	// Instagram
	InstagramPassword ConfigItem
	InstagramUsername ConfigItem

	// Mongo
	MongoHost     ConfigItem
	MongoPort     ConfigItem
	MongoUsername ConfigItem
	MongoPassword ConfigItem
	MongoDatabase ConfigItem

	// MySQL
	MySQLHost     ConfigItem
	MySQLPort     ConfigItem
	MySQLUsername ConfigItem
	MySQLPassword ConfigItem
	MySQLDatabase ConfigItem

	// Patreon
	PatreonSecret       ConfigItem
	PatreonClientID     ConfigItem
	PatreonClientSecret ConfigItem

	// Rabbit
	RabbitUsername      ConfigItem
	RabbitPassword      ConfigItem
	RabbitHost          ConfigItem
	RabbitPort          ConfigItem
	RabbitManagmentPort ConfigItem

	// Reddit
	RedditUsername ConfigItem
	RedditPassword ConfigItem

	// Recaptcha
	RecaptchaPrivate ConfigItem
	RecaptchaPublic  ConfigItem

	// Rollbar
	RollbarSecret ConfigItem
	RollbarUser   ConfigItem

	// Sentry
	SentryDSN ConfigItem

	// Sessions
	SessionAuthentication ConfigItem
	SessionEncryption     ConfigItem

	// Steam
	SteamUsername ConfigItem
	SteamPassword ConfigItem
	SteamAPIKey   ConfigItem // Not set from ENV

	// Twitch
	TwitchClientID     ConfigItem
	TwitchClientSecret ConfigItem

	// Twitter
	TwitterAccessToken       ConfigItem
	TwitterAccessTokenSecret ConfigItem
	TwitterConsumerKey       ConfigItem
	TwitterConsumerSecret    ConfigItem

	// Other
	Environment        ConfigItem
	GameDBDomain       ConfigItem
	GameDBShortName    ConfigItem
	GithubToken        ConfigItem
	MemcacheDSN        ConfigItem
	SendGridAPIKey     ConfigItem
	WebserverPort      ConfigItem
	CommitHash         ConfigItem
	NewReleaseDays     ConfigItem
	SlackGameDBWebhook ConfigItem
	SlackSocialWebhook ConfigItem
}

func init() {

	// Admin
	Config.AdminName.Set("ADMIN_NAME")
	Config.AdminEmail.Set("ADMIN_EMAIL")
	Config.AdminSteamID.Set("ADMIN_STEAM_ID")

	// Discord
	Config.DiscordClientID.Set("DISCORD_CLIENT_ID")
	Config.DiscordClientSescret.Set("DISCORD_SECRET")

	Config.DiscordChatBotToken.Set("DISCORD_BOT_TOKEN")
	Config.DiscordRelayBotToken.Set("DISCORD_RELAY_TOKEN")
	Config.DiscordChangesBotToken.Set("DISCORD_CHANGES_BOT_TOKEN")

	// Facebook
	Config.FacebookAppID.Set("FACEBOOK_APP_ID")
	Config.FacebookAppID.Set("FACEBOOK_APP_SECRET")

	// Google
	Config.GoogleBucket.Set("GOOGLE_BUCKET")
	Config.GoogleProject.Set("GOOGLE_PROJECT")

	Config.GoogleOauthClientID.Set("GOOGLE_OAUTH_CLIENT_ID")
	Config.GoogleOauthClientSecret.Set("GOOGLE_OAUTH_CLIENT_SECRET")

	// Influx
	Config.InfluxURL.Set("INFLUX_URL")
	Config.InfluxUsername.Set("INFLUX_USERNAME")
	Config.InfluxPassword.Set("INFLUX_PASSWORD")

	// Instagram
	Config.InstagramUsername.Set("INSTAGRAM_USERNAME")
	Config.InstagramPassword.Set("INSTAGRAM_PASSWORD")

	// Mongo
	Config.MongoHost.Set("MONGO_HOST")
	Config.MongoPort.Set("MONGO_PORT")
	Config.MongoUsername.Set("MONGO_USERNAME")
	Config.MongoPassword.Set("MONGO_PASSWORD")
	Config.MongoDatabase.Set("MONGO_DATABASE")

	// MySQL
	Config.MySQLHost.Set("MYSQL_HOST")
	Config.MySQLPort.Set("MYSQL_PORT")
	Config.MySQLUsername.Set("MYSQL_USERNAME")
	Config.MySQLPassword.Set("MYSQL_PASSWORD")
	Config.MySQLDatabase.Set("MYSQL_DATABASE")

	// Patreon
	Config.PatreonSecret.Set("PATREON_WEBOOK_SECRET")
	Config.PatreonClientID.Set("PATREON_CLIENT_ID")
	Config.PatreonClientSecret.Set("PATREON_CLIENT_SECRET")

	// Rabbit
	Config.RabbitUsername.Set("RABBIT_USER")
	Config.RabbitPassword.Set("RABBIT_PASS")
	Config.RabbitHost.Set("RABBIT_HOST")
	Config.RabbitPort.Set("RABBIT_PORT")
	Config.RabbitManagmentPort.Set("RABBIT_MANAGEMENT_PORT")

	// Reddit
	Config.RedditUsername.Set("REDDIT_USERNAME")
	Config.RedditPassword.Set("REDDIT_PASSWORD")

	// Recaptcha
	Config.RecaptchaPublic.Set("RECAPTCHA_PUBLIC")
	Config.RecaptchaPrivate.Set("RECAPTCHA_PRIVATE")

	// Rollbar
	Config.RollbarSecret.Set("ROLLBAR_PRIVATE")
	Config.RollbarUser.Set("ROLLBAR_USER")

	// Sentry
	Config.SentryDSN.Set("SENTRY_DSN")

	// Session
	Config.SessionAuthentication.Set("SESSION_AUTHENTICATION")
	Config.SessionEncryption.Set("SESSION_ENCRYPTION")

	// Steam
	Config.SteamUsername.Set("PROXY_USERNAME")
	Config.SteamPassword.Set("PROXY_PASSWORD")

	// Twitch
	Config.TwitchClientID.Set("TWITCH_CLIENT_ID")
	Config.TwitchClientSecret.Set("TWITCH_CLIENT_SECRET")

	// Twitter
	Config.TwitterAccessToken.Set("TWITTER_ACCESS_TOKEN")
	Config.TwitterAccessTokenSecret.Set("TWITTER_ACCESS_TOKEN_SECRET")
	Config.TwitterConsumerKey.Set("TWITTER_CONSUMER_KEY")
	Config.TwitterConsumerSecret.Set("TWITTER_CONSUMER_SECRET")

	// Other
	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.GithubToken.Set("GITHUB_TOKEN")
	Config.SendGridAPIKey.Set("SENDGRID")
	Config.WebserverPort.Set("PORT")
	Config.SlackGameDBWebhook.Set("SLACK_GAMEDB_WEBHOOK")
	Config.SlackSocialWebhook.Set("SLACK_SOCIAL_WEBHOOK")

	// Defaults
	Config.GameDBShortName.SetDefault("GameDB")
	Config.NewReleaseDays.SetDefault("14")
	Config.SteamAPIKey.SetDefault("unset")

	switch Config.Environment.Get() {
	case EnvProd:

		Config.MemcacheDSN.SetDefault("memcache:11211")

	case EnvLocal:

		Config.MemcacheDSN.SetDefault("localhost:11211")

	case EnvConsumer:

		Config.MemcacheDSN.SetDefault("memcache:11211")

	default:
		fmt.Println("config: unknown environment: " + Config.Environment.Get())
		os.Exit(1)
	}
}

// ConfigItem
type ConfigItem struct {
	value        string
	defaultValue string
}

func (ci *ConfigItem) Set(environment string) {
	ci.value = os.Getenv(prefix + environment)
}

func (ci *ConfigItem) SetDefault(defaultValue string) {
	ci.defaultValue = defaultValue
}

func (ci ConfigItem) Get() string {

	if ci.value != "" {
		return ci.value
	}

	if ci.defaultValue != "" {
		return ci.defaultValue
	}

	// Get line number of calling function
	skip := 1
	for {
		_, file, no, ok := runtime.Caller(skip)
		if ok {
			if !strings.Contains(file, "config/config.go") {
				fmt.Printf("missing env var @ %s#%d\n", file, no)
				break
			}
		} else if !ok || skip > 10 {
			fmt.Println("missing env var...")
			break
		}
		skip++
	}

	return ""
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

//
func RabbitDSN() string {
	return "amqp://" + Config.RabbitUsername.Get() + ":" + Config.RabbitPassword.Get() + "@" + Config.RabbitHost.Get() + ":" + Config.RabbitPort.Get()
}

func MySQLDNS() string {
	return Config.MySQLUsername.Get() + ":" + Config.MySQLPassword.Get() + "@tcp(" + Config.MySQLHost.Get() + ":" + Config.MySQLPort.Get() + ")/" + Config.MySQLDatabase.Get()
}

func MongoDSN() string {
	return "mongodb://" + Config.MongoHost.Get() + ":" + Config.MongoPort.Get()
}

// func RabbitAPI(values url.Values) string {
// 	return "http://" + Config.RabbitHost.Get() + ":" + Config.RabbitManagmentPort.Get() + "/api/overview?" + values.Encode()
// }

func ListenOn() string {
	return "0.0.0.0:" + Config.WebserverPort.Get()
}

func IsLocal() bool {
	return Config.Environment.Get() == EnvLocal
}

func IsProd() bool {
	return Config.Environment.Get() == EnvProd
}

func GetSteamKeyTag() string {

	key := Config.SteamAPIKey.Get()
	if len(key) > 7 {
		key = key[0:7]
	}

	return strings.ToUpper(key)
}

func SetVersion(v string) {

	if IsLocal() && v == "" {
		v = "local"
	}
	Config.CommitHash.SetDefault(v)
}

func GetShortVersion() string {

	key := Config.CommitHash.Get()
	if len(key) > 7 {
		key = key[0:7]
	}
	return key
}
