package config

import (
	"errors"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

const (
	EnvProd     = "production"
	EnvLocal    = "local"
	EnvConsumer = "consumer"
)

var ErrMissingEnvironmentVariable = errors.New("missing env var")

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
	DiscordClientID        string `envconfig:"DISCORD_CLIENT_ID"`         // OAuth
	DiscordClientSescret   string `envconfig:"DISCORD_SECRET"`            // OAuth
	DiscordChatBotToken    string `envconfig:"DISCORD_BOT_TOKEN"`         // Bot
	DiscordRelayBotToken   string `envconfig:"DISCORD_RELAY_TOKEN"`       // Bot
	DiscordChangesBotToken string `envconfig:"DISCORD_CHANGES_BOT_TOKEN"` // Bot
	DiscordOAuthBotToken   string `envconfig:"DISCORD_OAUTH_BOT_TOKEN"`   // Bot

	// Elastic Search
	ElasticAddress  string `envconfig:"ELASTIC_SEARCH_ADDRESS" required:"true"`
	ElasticUsername string `envconfig:"ELASTIC_SEARCH_USERNAME" required:"true"`
	ElasticPassword string `envconfig:"ELASTIC_SEARCH_PASSWORD"`

	// GitHub
	GitHubClient        string `envconfig:"GITHUB_CLIENT"`         // OAuth
	GitHubSecret        string `envconfig:"GITHUB_SECRET"`         // OAuth
	GithubToken         string `envconfig:"GITHUB_TOKEN"`          // API
	GithubWebhookSecret string `envconfig:"GITHUB_WEBHOOK_SECRET"` // Webhooks

	// Google
	GoogleOauthClientID     string `envconfig:"GOOGLE_OAUTH_CLIENT_ID"`     // OAuth
	GoogleOauthClientSecret string `envconfig:"GOOGLE_OAUTH_CLIENT_SECRET"` // OAuth
	GoogleProject           string `envconfig:"GOOGLE_PROJECT"`             // Logging
	GoogleAuthFile          string `envconfig:"GOOGLE_AUTH"`                // Logging

	// Hetzner
	HetznerSSHKeyID  int    `envconfig:"HETZNER_SSH_KEY_ID"` // Scaler
	HetznerNetworkID int    `envconfig:"HETZNER_NETWORK_ID"` // Scaler
	HetznerAPIToken  string `envconfig:"HETZNER_API_TOKEN"`  // Scaler

	// Influx
	InfluxURL      string `envconfig:"INFLUX_URL" required:"true"`
	InfluxUsername string `envconfig:"INFLUX_USERNAME" required:"true"`
	InfluxPassword string `envconfig:"INFLUX_PASSWORD"`

	// Instagram
	InstagramUsername string `envconfig:"INSTAGRAM_USERNAME"`
	InstagramPassword string `envconfig:"INSTAGRAM_PASSWORD"`

	// Mailjet
	MailjetPublic  string `envconfig:"MAILJET_PUBLIC" required:"true"`  // API
	MailjetPrivate string `envconfig:"MAILJET_PRIVATE" required:"true"` // API

	// Memcache
	MemcacheDSN      string `envconfig:"MEMCACHE_URL" required:"true"`
	MemcacheUsername string `envconfig:"MEMCACHE_USERNAME"`
	MemcachePassword string `envconfig:"MEMCACHE_PASSWORD"`

	// Mongo
	MongoHost     string `envconfig:"MONGO_HOST" required:"true"`
	MongoPort     string `envconfig:"MONGO_PORT" required:"true"`
	MongoUsername string `envconfig:"MONGO_USERNAME" required:"true"`
	MongoPassword string `envconfig:"MONGO_PASSWORD"`
	MongoDatabase string `envconfig:"MONGO_DATABASE" required:"true"`

	// MySQL
	MySQLHost     string `envconfig:"MYSQL_HOST" required:"true"`
	MySQLPort     string `envconfig:"MYSQL_PORT" required:"true"`
	MySQLUsername string `envconfig:"MYSQL_USERNAME" required:"true"`
	MySQLPassword string `envconfig:"MYSQL_PASSWORD"`
	MySQLDatabase string `envconfig:"MYSQL_DATABASE" required:"true"`

	// Patreon
	PatreonSecret       string `envconfig:"PATREON_WEBOOK_SECRET"` // Webhooks
	PatreonClientID     string `envconfig:"PATREON_CLIENT_ID"`     // OAuth
	PatreonClientSecret string `envconfig:"PATREON_CLIENT_SECRET"` // OAuth

	// Rabbit
	RabbitUsername      string `envconfig:"RABBIT_USER" required:"true"`
	RabbitPassword      string `envconfig:"RABBIT_PASS" required:"true"`
	RabbitHost          string `envconfig:"RABBIT_HOST" required:"true"`
	RabbitPort          string `envconfig:"RABBIT_PORT" required:"true"`
	RabbitManagmentPort string `envconfig:"RABBIT_MANAGEMENT_PORT" required:"true"`

	// Recaptcha
	RecaptchaPublic  string `envconfig:"RECAPTCHA_PUBLIC"`
	RecaptchaPrivate string `envconfig:"RECAPTCHA_PRIVATE"`

	// Rollbar
	RollbarSecret string `envconfig:"ROLLBAR_PRIVATE"`
	RollbarUser   string `envconfig:"ROLLBAR_USER"`

	// Sendgrid
	SendGridSecret string `envconfig:"SENDGRID_WEBHOOK_SECRET"`
	SendGridAPIKey string `envconfig:"SENDGRID"`

	// Sentry
	SentryDSN string `envconfig:"SENTRY_DSN"`

	// Session
	SessionAuthentication string `envconfig:"SESSION_AUTHENTICATION" required:"true"`
	SessionEncryption     string `envconfig:"SESSION_ENCRYPTION" required:"true"`

	// Steam
	SteamUsername string `envconfig:"PROXY_USERNAME"`
	SteamPassword string `envconfig:"PROXY_PASSWORD"`
	SteamAPIKey   string

	// Twitch
	TwitchClientID     string `envconfig:"TWITCH_CLIENT_ID"`
	TwitchClientSecret string `envconfig:"TWITCH_CLIENT_SECRET"`

	// Twitter
	TwitterAccessToken       string `envconfig:"TWITTER_ACCESS_TOKEN"`        // API (Home)
	TwitterAccessTokenSecret string `envconfig:"TWITTER_ACCESS_TOKEN_SECRET"` // API (Home)
	TwitterConsumerKey       string `envconfig:"TWITTER_CONSUMER_KEY"`        // API (Home)
	TwitterConsumerSecret    string `envconfig:"TWITTER_CONSUMER_SECRET"`     // API (Home)
	TwitterZapierSecret      string `envconfig:"TWITTER_ZAPIER_SECRET"`       // Webhooks

	// YouTube
	YoutubeAPIKey string `envconfig:"YOUTUBE_API_KEY"`

	// Servers
	FrontendPort      string `envconfig:"PORT"`
	BackendHostPort   string `envconfig:"BACKEND_HOST_PORT"`
	BackendClientPort string `envconfig:"BACKEND_CLIENT_PORT"`
	APIPort           string `envconfig:"API_PORT"`

	// Other
	GameDBDomain        string `envconfig:"DOMAIN"` // With proto & port
	Environment         string `envconfig:"ENV" required:"true"`
	SlackGameDBWebhook  string `envconfig:"SLACK_GAMEDB_WEBHOOK"`
	SlackPatreonWebhook string `envconfig:"SLACK_SOCIAL_WEBHOOK"`
	ChatBotAttachments  string `envconfig:"CHATBOT_ATTACHMENTS"`
	GRPCKeysPath        string `envconfig:"GRPC_KEYS_PATH"`

	// Non-environment
	CommitHash      string `ignored:"true"`
	Commits         string `ignored:"true"`
	GameDBShortName string `ignored:"true"`
	IP              string `ignored:"true"`
	NewReleaseDays  int    `ignored:"true"`
}

var C Config

func Init(version string, commits string, ip string) (err error) {

	err = envconfig.Process("steam", &C)

	C.CommitHash = version
	C.Commits = commits
	C.GameDBShortName = "GameDB"
	C.IP = ip
	C.NewReleaseDays = 14

	return err
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
