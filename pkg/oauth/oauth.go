package oauth

import (
	"net/http"

	"github.com/dghubble/oauth1"
	"golang.org/x/oauth2"
)

type ProviderType int

const (
	TypeOAuth ProviderType = iota
	TypeOpenID
)

type ProviderEnum string

var (
	ProviderDiscord   ProviderEnum = "discord"
	ProviderGoogle    ProviderEnum = "google"
	ProviderGithub    ProviderEnum = "github"
	ProviderBattlenet ProviderEnum = "battlenetus"
	ProviderPatreon   ProviderEnum = "patreon"
	ProviderSteam     ProviderEnum = "steam"
	ProviderTwitter   ProviderEnum = "twitter"
)

var Providers = []Provider{
	&steamProvider{},
	&discordProvider{},
	&battlenetProvider{},
	&googleProvider{},
	&twitterProvider{},
	&patreonProvider{},
	&githubProvider{},
}

var _ = []OAuth1Provider{
	twitterProvider{},
}

var _ = []OpenIDProvider{
	steamProvider{},
}

type Provider interface {
	GetName() string
	GetIcon() string
	GetColour() string
	GetEnum() ProviderEnum
	GetType() ProviderType
	HasEmail() bool
}

type OAuth2Provider interface {
	Provider
	GetConfig() oauth2.Config
	GetUser(*oauth2.Token) (User, error)
	Redirect(http.ResponseWriter, *http.Request, string)
}

type OAuth1Provider interface {
	Provider
	GetConfig() oauth1.Config
	GetUser(*oauth1.Token) (User, error)
	Redirect() (string, string, error)
}

type OpenIDProvider interface {
	Provider
	GetUser(*http.Request) (User, error)
	Redirect(http.ResponseWriter, *http.Request, string)
}

func New(p ProviderEnum) Provider {

	for _, v := range Providers {
		if p == v.GetEnum() {
			return v
		}
	}

	return nil
}

//
type User struct {
	Token    string
	ID       string
	Username string
	Email    string
	Avatar   string
}
