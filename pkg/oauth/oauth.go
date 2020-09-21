package oauth

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/oauth2"
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

type Provider interface {
	GetName() string
	GetIcon() string
	GetColour() string
	GetEnum() ProviderEnum
	Redirect(w http.ResponseWriter, r *http.Request, state string)
	GetUser(r *http.Request, token *oauth2.Token) (User, error) // r for OpenID, token for OAuth
}

type OAuthProvider interface {
	Provider
	GetConfig() oauth2.Config
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

func (u User) IDInt() int {
	i, err := strconv.Atoi(u.ID)
	if err != nil {
		log.Err(err.Error())
	}
	return i
}

//
type State struct {
	State string
	Page  string
}

func (s State) Marshal() string {
	b, err := json.Marshal(s)
	if err != nil {
		log.ErrS(err)
	}
	return string(b)
}

func (s *State) Unmarshal(in string) {
	err := json.Unmarshal([]byte(in), s)
	if err != nil {
		log.ErrS(err)
	}
}
