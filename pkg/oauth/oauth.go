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
	ProviderDiscord ProviderEnum = "discord"
	ProviderGoogle  ProviderEnum = "google"
	ProviderGithub  ProviderEnum = "github"
	ProviderPatreon ProviderEnum = "patreon"
	ProviderSteam   ProviderEnum = "steam"
	ProviderTwitter ProviderEnum = "twitter"
)

var Providers = []Provider{
	discordProvider{},
	googleProvider{},
	patreonProvider{},
	steamProvider{},
	githubProvider{},
	twitterProvider{},
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

	switch p {
	case ProviderDiscord:
		return discordProvider{}
	case ProviderGoogle:
		return googleProvider{}
	case ProviderPatreon:
		return patreonProvider{}
	case ProviderSteam:
		return steamProvider{}
	case ProviderGithub:
		return githubProvider{}
	case ProviderTwitter:
		return twitterProvider{}
	default:
		return nil
	}
}

//
type OauthError struct {
	Err   error
	Flash string
}

func (oe OauthError) Error() string {
	if oe.Err != nil {
		return oe.Err.Error()
	}
	return ""
}

//
type User struct {
	Token    string
	ID       string
	Username string
	Email    string
	Avatar   string
}
