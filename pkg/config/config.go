package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

//noinspection GoUnusedConst
const (
	EnvProd     = "production"
	EnvLocal    = "local"
	EnvConsumer = "consumer"
)

var Config BaseConfig

type BaseConfig struct {
	// Admin
	AdminName    ConfigItem
	AdminEmail   ConfigItem
	AdminSteamID ConfigItem

	// Digital Ocean (Auto Scaler)
	DigitalOceanAccessToken    ConfigItem
	DigitalOceanProjectID      ConfigItem
	DigitalOceanKeyID          ConfigItem
	DigitalOceanKeyFingerprint ConfigItem

	// Discord
	DiscordClientID      ConfigItem
	DiscordClientSescret ConfigItem

	DiscordRelayBotToken   ConfigItem
	DiscordChatBotToken    ConfigItem
	DiscordChangesBotToken ConfigItem

	// Elastic Seach
	ElasticAddress  ConfigItem
	ElasticUsername ConfigItem
	ElasticPassword ConfigItem

	// Facebook
	FacebookAppID     ConfigItem
	FacebookAppSecret ConfigItem

	// GitHub
	GitHubClient        ConfigItem
	GitHubSecret        ConfigItem
	GithubToken         ConfigItem
	GithubWebhookSecret ConfigItem

	// Google
	GoogleBucket  ConfigItem
	GoogleProject ConfigItem

	GoogleOauthClientID     ConfigItem
	GoogleOauthClientSecret ConfigItem

	// Hetzner (Auto Scaler)
	HetznerSSHKeyID  ConfigItem
	HetznerNetworkID ConfigItem
	HetznerAPIToken  ConfigItem

	// Imgur
	ImgurClientID ConfigItem

	// Influx
	InfluxURL      ConfigItem
	InfluxPassword ConfigItem
	InfluxUsername ConfigItem

	// Instagram
	InstagramPassword ConfigItem
	InstagramUsername ConfigItem

	// Memcache
	MemcacheDSN      ConfigItem
	MemcacheUsername ConfigItem
	MemcachePassword ConfigItem

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

	// Versions
	CommitHash ConfigItem
	Commits    ConfigItem

	// YouTube
	YoutubeAPIKey ConfigItem

	// Servers
	FrontendPort ConfigItem
	APIPort      ConfigItem

	// Other
	Environment         ConfigItem
	GameDBDomain        ConfigItem
	GameDBShortName     ConfigItem
	SendGridAPIKey      ConfigItem
	IP                  ConfigItem
	NewReleaseDays      ConfigItem
	SlackGameDBWebhook  ConfigItem
	SlackPatreonWebhook ConfigItem
	InfraPath           ConfigItem
}

func init() {

	// Admin
	Config.AdminName.Set("ADMIN_NAME")
	Config.AdminEmail.Set("ADMIN_EMAIL")
	Config.AdminSteamID.Set("ADMIN_STEAM_ID")

	// Digital Ocean (Auto Scaler)
	Config.DigitalOceanAccessToken.Set("DO_ACCESS_TOKEN")
	Config.DigitalOceanProjectID.Set("DO_PROJECT_ID")
	Config.DigitalOceanKeyID.Set("DO_KEY_ID")
	Config.DigitalOceanKeyFingerprint.Set("DO_KEY_FINGERPRINT")

	// Discord
	Config.DiscordClientID.Set("DISCORD_CLIENT_ID")
	Config.DiscordClientSescret.Set("DISCORD_SECRET")

	Config.DiscordChatBotToken.Set("DISCORD_BOT_TOKEN")
	Config.DiscordRelayBotToken.Set("DISCORD_RELAY_TOKEN")
	Config.DiscordChangesBotToken.Set("DISCORD_CHANGES_BOT_TOKEN")

	// Elastic Search
	Config.ElasticAddress.Set("ELASTIC_SEARCH_ADDRESS")
	Config.ElasticUsername.Set("ELASTIC_SEARCH_USERNAME")
	Config.ElasticPassword.Set("ELASTIC_SEARCH_PASSWORD")

	// Facebook
	Config.FacebookAppID.Set("FACEBOOK_APP_ID")
	Config.FacebookAppID.Set("FACEBOOK_APP_SECRET")

	// GitHub
	Config.GitHubClient.Set("GITHUB_CLIENT")
	Config.GitHubSecret.Set("GITHUB_SECRET")
	Config.GithubToken.Set("GITHUB_TOKEN")
	Config.GithubWebhookSecret.Set("GITHUB_WEBHOOK_SECRET")

	// Google
	Config.GoogleBucket.Set("GOOGLE_BUCKET")
	Config.GoogleProject.Set("GOOGLE_PROJECT")

	Config.GoogleOauthClientID.Set("GOOGLE_OAUTH_CLIENT_ID")
	Config.GoogleOauthClientSecret.Set("GOOGLE_OAUTH_CLIENT_SECRET")

	// Hetzner
	Config.HetznerSSHKeyID.Set("HETZNER_SSH_KEY_ID")
	Config.HetznerNetworkID.Set("HETZNER_NETWORK_ID")
	Config.HetznerAPIToken.Set("HETZNER_API_TOKEN")

	// Imgur
	Config.ImgurClientID.Set("IMGUR_CLIENT_ID")

	// Influx
	Config.InfluxURL.Set("INFLUX_URL")
	Config.InfluxUsername.Set("INFLUX_USERNAME")
	Config.InfluxPassword.Set("INFLUX_PASSWORD")

	// Instagram
	Config.InstagramUsername.Set("INSTAGRAM_USERNAME")
	Config.InstagramPassword.Set("INSTAGRAM_PASSWORD")

	// Memcache
	Config.MemcacheDSN.Set("MEMCACHE_URL")
	Config.MemcacheUsername.Set("MEMCACHE_USERNAME", true)
	Config.MemcachePassword.Set("MEMCACHE_PASSWORD", true)

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

	// YouTube
	Config.YoutubeAPIKey.Set("YOUTUBE_API_KEY")

	// Servers
	Config.FrontendPort.Set("PORT")
	Config.APIPort.Set("API_PORT")

	// Other
	Config.GameDBDomain.Set("DOMAIN")
	Config.Environment.Set("ENV")
	Config.SendGridAPIKey.Set("SENDGRID")
	Config.SlackGameDBWebhook.Set("SLACK_GAMEDB_WEBHOOK")
	Config.SlackPatreonWebhook.Set("SLACK_SOCIAL_WEBHOOK")
	Config.InfraPath.Set("INFRASTRUCTURE_PATH")

	// Defaults
	Config.GameDBShortName.SetDefault("GameDB")
	Config.NewReleaseDays.SetDefault("14")
	Config.SteamAPIKey.SetDefault("")
}

func Init(version string, commits string, ip string) {
	Config.CommitHash.SetDefault(version)
	Config.Commits.SetDefault(commits)
	Config.IP.SetDefault(ip)
}

// ConfigItem
type ConfigItem struct {
	value        string
	defaultValue string
	allowEmpty   bool
}

func (ci *ConfigItem) Set(environment string, allowEmpty ...bool) *ConfigItem {

	ci.value = os.Getenv("STEAM_" + environment)
	ci.allowEmpty = len(allowEmpty) > 0 && allowEmpty[0]
	return ci
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

	// Log line number of config item
	if !ci.allowEmpty {
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

func (ci ConfigItem) GetInt64() int64 {
	i, err := strconv.ParseInt(ci.Get(), 10, 64)
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

func FrontendPort() string {
	return "0.0.0.0:" + Config.FrontendPort.Get()
}

func APIPort() string {
	return "0.0.0.0:" + Config.APIPort.Get()
}

func IsLocal() bool {
	return Config.Environment.Get() == EnvLocal
}

func IsProd() bool {
	return Config.Environment.Get() == EnvProd
}

func IsConsumer() bool {
	return Config.Environment.Get() == EnvConsumer
}

func GetSteamKeyTag() string {

	key := Config.SteamAPIKey.Get()
	if len(key) > 7 {
		key = key[0:7]
	}

	return strings.ToUpper(key)
}

func GetShortCommitHash() string {

	key := Config.CommitHash.Get()
	if len(key) > 7 {
		key = key[0:7]
	}
	return key
}
