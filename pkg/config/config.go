package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

//noinspection GoUnusedConst
const (
	EnvProd     = "production"
	EnvLocal    = "local"
	EnvConsumer = "consumer"
)

type Config struct {

	// Admin
	AdminName  string `envconfig:"ADMIN_NAME"`
	AdminEmail string `envconfig:"ADMIN_EMAIL"`

	// Digital Ocean (Auto Scaler)
	DigitalOceanAccessToken    string `envconfig:"DO_ACCESS_TOKEN"`
	DigitalOceanProjectID      string `envconfig:"DO_PROJECT_ID"`
	DigitalOceanKeyID          int    `envconfig:"DO_KEY_ID"`
	DigitalOceanKeyFingerprint string `envconfig:"DO_KEY_FINGERPRINT"`

	// Discord
	DiscordClientID      string `envconfig:"DISCORD_CLIENT_ID"`
	DiscordClientSescret string `envconfig:"DISCORD_SECRET"`

	DiscordChatBotToken    string `envconfig:"DISCORD_BOT_TOKEN"`
	DiscordRelayBotToken   string `envconfig:"DISCORD_RELAY_TOKEN"`
	DiscordChangesBotToken string `envconfig:"DISCORD_CHANGES_BOT_TOKEN"`

	// Elastic Search
	ElasticAddress  string `envconfig:"ELASTIC_SEARCH_ADDRESS"`
	ElasticUsername string `envconfig:"ELASTIC_SEARCH_USERNAME"`
	ElasticPassword string `envconfig:"ELASTIC_SEARCH_PASSWORD"`

	// GitHub
	GitHubClient        string `envconfig:"GITHUB_CLIENT"`
	GitHubSecret        string `envconfig:"GITHUB_SECRET"`
	GithubToken         string `envconfig:"GITHUB_TOKEN"`
	GithubWebhookSecret string `envconfig:"GITHUB_WEBHOOK_SECRET"`

	// Google
	GoogleProject           string `envconfig:"GOOGLE_PROJECT"`
	GoogleOauthClientID     string `envconfig:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOauthClientSecret string `envconfig:"GOOGLE_OAUTH_CLIENT_SECRET"`

	// Hetzner
	HetznerSSHKeyID  int    `envconfig:"HETZNER_SSH_KEY_ID"`
	HetznerNetworkID int    `envconfig:"HETZNER_NETWORK_ID"`
	HetznerAPIToken  string `envconfig:"HETZNER_API_TOKEN"`

	// Influx
	InfluxURL      string `envconfig:"INFLUX_URL"`
	InfluxUsername string `envconfig:"INFLUX_USERNAME"`
	InfluxPassword string `envconfig:"INFLUX_PASSWORD"`

	// Instagram
	InstagramUsername string `envconfig:"INSTAGRAM_USERNAME"`
	InstagramPassword string `envconfig:"INSTAGRAM_PASSWORD"`

	// Memcache
	MemcacheDSN      string `envconfig:"MEMCACHE_URL"`
	MemcacheUsername string `envconfig:"MEMCACHE_USERNAME"`
	MemcachePassword string `envconfig:"MEMCACHE_PASSWORD"`

	// Mongo
	MongoHost     string `envconfig:"MONGO_HOST"`
	MongoPort     string `envconfig:"MONGO_PORT"`
	MongoUsername string `envconfig:"MONGO_USERNAME"`
	MongoPassword string `envconfig:"MONGO_PASSWORD"`
	MongoDatabase string `envconfig:"MONGO_DATABASE"`

	// MySQL
	MySQLHost     string `envconfig:"MYSQL_HOST"`
	MySQLPort     string `envconfig:"MYSQL_PORT"`
	MySQLUsername string `envconfig:"MYSQL_USERNAME"`
	MySQLPassword string `envconfig:"MYSQL_PASSWORD"`
	MySQLDatabase string `envconfig:"MYSQL_DATABASE"`

	// Patreon
	PatreonSecret       string `envconfig:"PATREON_WEBOOK_SECRET"`
	PatreonClientID     string `envconfig:"PATREON_CLIENT_ID"`
	PatreonClientSecret string `envconfig:"PATREON_CLIENT_SECRET"`

	// Rabbit
	RabbitUsername      string `envconfig:"RABBIT_USER"`
	RabbitPassword      string `envconfig:"RABBIT_PASS"`
	RabbitHost          string `envconfig:"RABBIT_HOST"`
	RabbitPort          string `envconfig:"RABBIT_PORT"`
	RabbitManagmentPort string `envconfig:"RABBIT_MANAGEMENT_PORT"`

	// Recaptcha
	RecaptchaPublic  string `envconfig:"RECAPTCHA_PUBLIC"`
	RecaptchaPrivate string `envconfig:"RECAPTCHA_PRIVATE"`

	// Rollbar
	RollbarSecret string `envconfig:"ROLLBAR_PRIVATE"`
	RollbarUser   string `envconfig:"ROLLBAR_USER"`

	// Sentry
	SentryDSN string `envconfig:"SENTRY_DSN"`

	// Session
	SessionAuthentication string `envconfig:"SESSION_AUTHENTICATION"`
	SessionEncryption     string `envconfig:"SESSION_ENCRYPTION"`

	// Steam
	SteamUsername string `envconfig:"PROXY_USERNAME"`
	SteamPassword string `envconfig:"PROXY_PASSWORD"`
	SteamAPIKey   string

	// Twitch
	TwitchClientID     string `envconfig:"TWITCH_CLIENT_ID"`
	TwitchClientSecret string `envconfig:"TWITCH_CLIENT_SECRET"`

	// Twitter
	TwitterAccessToken       string `envconfig:"TWITTER_ACCESS_TOKEN"`
	TwitterAccessTokenSecret string `envconfig:"TWITTER_ACCESS_TOKEN_SECRET"`
	TwitterConsumerKey       string `envconfig:"TWITTER_CONSUMER_KEY"`
	TwitterConsumerSecret    string `envconfig:"TWITTER_CONSUMER_SECRET"`

	// YouTube
	YoutubeAPIKey string `envconfig:"YOUTUBE_API_KEY"`

	// Servers
	FrontendPort      string `envconfig:"PORT"`
	BackendHostPort   string `envconfig:"BACKEND_HOST_PORT"`
	BackendClientPort string `envconfig:"BACKEND_CLIENT_PORT"`
	APIPort           string `envconfig:"API_PORT"`

	// Other
	GameDBDomain        string `envconfig:"DOMAIN"` // With proto & port
	Environment         string `envconfig:"ENV"`
	SendGridAPIKey      string `envconfig:"SENDGRID"`
	SlackGameDBWebhook  string `envconfig:"SLACK_GAMEDB_WEBHOOK"`
	SlackPatreonWebhook string `envconfig:"SLACK_SOCIAL_WEBHOOK"`
	InfraPath           string `envconfig:"INFRASTRUCTURE_PATH"`
	ChatBotAttachments  string `envconfig:"CHATBOT_ATTACHMENTS"`
	GRPCKeysPath        string `envconfig:"GRPC_KEYS_PATH"`

	// Non-environment
	IP              string
	CommitHash      string
	Commits         string
	GameDBShortName string
	NewReleaseDays  int
}

var C Config

func init() {

	err := envconfig.Process("steam", &C)
	if err != nil {
		fmt.Println(err) // Zap not ready yet
	}

	C.GameDBShortName = "GameDB"
	C.NewReleaseDays = 14
}

func Init(version string, commits string, ip string) {
	C.CommitHash = version
	C.Commits = commits
	C.IP = ip
}

func MySQLDNS() string {
	return C.MySQLUsername + ":" + C.MySQLPassword + "@tcp(" + C.MySQLHost + ":" + C.MySQLPort + ")/" + C.MySQLDatabase
}

func RabbitDSN() string {
	return "amqp://" + C.RabbitUsername + ":" + C.RabbitPassword + "@" + C.RabbitHost + ":" + C.RabbitPort
}

func MongoDSN() string {
	return "mongodb://" + C.MongoHost + ":" + C.MongoPort
}

func GetFrontendPort() string {
	return "0.0.0.0:" + C.FrontendPort
}

func GetAPIPort() string {
	return "0.0.0.0:" + C.APIPort
}

func IsLocal() bool {
	return C.Environment == EnvLocal
}

func IsProd() bool {
	return C.Environment == EnvProd
}

func IsConsumer() bool {
	return C.Environment == EnvConsumer
}

func GetSteamKeyTag() string {

	key := C.SteamAPIKey
	if len(key) > 7 {
		key = key[0:7]
	}

	return strings.ToUpper(key)
}

func GetShortCommitHash() string {

	key := C.CommitHash
	if len(key) > 7 {
		key = key[0:7]
	}
	return key
}
