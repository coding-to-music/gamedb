package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
)

const (
	guildID          = "407493776597057538"
	generalChannelID = "407493777058693121"
	sessionKey       = "discord_token"
)

var (
	discordOauthContext = context.Background()
	discordOauthConfig  = &oauth2.Config{
		ClientID:     config.Config.DiscordClientID,
		ClientSecret: config.Config.DiscordSescret,
		Scopes:       []string{"identify", "guilds.join", "guilds"},
		RedirectURL:  "https://gamedb.online/chat/callback/",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discordapp.com/api/oauth2/authorize",
			TokenURL: "https://discordapp.com/api/oauth2/token",
		},
	}
)

func chatRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", chatHandler)
	r.Get("/login", chatLoginHandler)
	r.Get("/callback", chatLoginCallbackHandler)
	r.Get("/{id}", chatHandler)
	r.Post("/{id}/post", chatPostHandler)
	r.Get("/{id}/chat.json", chatAjaxHandler)
	return r
}

func chatHandler(w http.ResponseWriter, r *http.Request) {

	// Get ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Redirect(w, r, "/chat/"+generalChannelID, 302)
		return
	}

	// Template
	t := chatTemplate{}
	t.Fill(w, r, "Chat", "The Game DB community.")
	t.ChannelID = id
	t.addAssetJSON2HTML()

	//
	var wg sync.WaitGroup
	var discordErr error

	// Get channels
	wg.Add(1)
	go func() {

		defer wg.Done()

		var channelsResponse []*discordgo.Channel

		operation := func() (err error) {

			discord, err := getDiscord(r)
			if err != nil {
				return err
			}

			channelsResponse, err = discord.GuildChannels(guildID)
			return err
		}

		policy := backoff.NewExponentialBackOff()

		err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
		if err != nil {
			discordErr = err
			log.Critical(err, r)
		}

		for _, v := range channelsResponse {
			if v.Type == discordgo.ChannelTypeGuildText {

				// Fix channel name
				v.Name = strings.Replace(v.Name, "-", " ", 1)
				v.Name = strings.Replace(v.Name, "db", "DB", 1)
				v.Name = strings.Title(v.Name)

				t.Channels = append(t.Channels, v)
			}
		}

	}()

	// Get members
	wg.Add(1)
	go func() {

		defer wg.Done()

		var membersResponse []*discordgo.Member

		operation := func() (err error) {

			discord, err := getDiscord(r)
			if err != nil {
				return err
			}

			membersResponse, err = discord.GuildMembers(guildID, "", 1000)
			return err
		}

		policy := backoff.NewExponentialBackOff()

		err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
		if err != nil {
			discordErr = err
			log.Critical(err, r)
		}

		for _, v := range membersResponse {
			if !v.User.Bot {
				t.Members = append(t.Members, v)
			}
		}

	}()

	// Wait
	wg.Wait()

	if discordErr != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Could not connect to Discord."})
		return
	}

	err := returnTemplate(w, r, "chat", t)
	log.Err(err, r)
}

type chatTemplate struct {
	GlobalTemplate
	ChannelID string
	Channels  []*discordgo.Channel
	Members   []*discordgo.Member
}

func chatAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	id := chi.URLParam(r, "id")
	if id == "" {
		id = generalChannelID
	}

	var messagesResponse []*discordgo.Message

	operation := func() (err error) {

		discord, err := getDiscord(r)
		if err != nil {
			return err
		}

		messagesResponse, err = discord.ChannelMessages(id, 50, "", "", "")
		return err
	}

	policy := backoff.NewExponentialBackOff()

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	if err != nil {
		log.Critical(err, r)
		return
	}

	var messages []chatWebsocketPayload
	for _, v := range messagesResponse {
		if !v.Author.Bot && v.Type == discordgo.MessageTypeDefault {

			messages = append(messages, chatWebsocketPayload{
				AuthorID:     v.Author.ID,
				AuthorUser:   v.Author.Username,
				AuthorAvatar: v.Author.Avatar,
				Content:      v.Content,
				Channel:      v.ChannelID,
			})
		}
	}

	bytes, err := json.Marshal(messages)
	log.Err(err, r)

	err = returnJSON(w, r, bytes)
	log.Err(err, r)
}

type chatWebsocketPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorUser   string `json:"author_user"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
	Channel      string `json:"channel"`
}

func chatLoginHandler(w http.ResponseWriter, r *http.Request) {

	if config.Config.IsLocal() {
		discordOauthConfig.RedirectURL = config.Config.GameDBDomain.Get() + "/chat/callback/"
	}

	http.Redirect(w, r, discordOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOnline), 302)
}

func chatLoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	var code = r.URL.Query().Get("code")

	// Get token
	tok, err := discordOauthConfig.Exchange(discordOauthContext, code)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Error: err, Message: "Something went wrong logging you in.", Code: 400})
		return
	}

	// Add user to guild
	discord, err := getDiscord(r)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Error: err, Message: "Something went wrong logging you in.", Code: 400})
		return
	}

	err = discord.GuildMemberAdd(tok.AccessToken, guildID, "@me", "", []string{}, false, false)
	if err != nil {
		log.Err(err)
		returnErrorTemplate(w, r, errorTemplate{Error: err, Message: "Something went wrong logging you in.", Code: 400})
		return
	}

	b, err := json.Marshal(tok)

	err = session.Write(w, r, sessionKey, string(b))
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Error: err, Message: "Something went wrong logging you in."})
		return
	}

	http.Redirect(w, r, "/chat", 302)
}

func chatPostHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Posting")

	err := r.ParseForm()
	if err != nil {
		log.Err(err)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("Missing channel ID")
		return
	}

	discord, err := getDiscordOauth(r)
	if err != nil {
		log.Err(err)
		return
	}

	_, err = discord.ChannelMessageSend(id, r.FormValue("message"))
	if err != nil {
		log.Err(err)
		return
	}
}

func getDiscord(r *http.Request) (discord *discordgo.Session, err error) {

	return helpers.GetDiscord(discordMessageHandler)
}

func getDiscordOauth(r *http.Request) (discordDeref discordgo.Session, err error) {

	discord, err := getDiscord(r)
	if err != nil {
		return
	}

	discordDeref = *discord

	// Add oauth2 http client to discord
	tokenString, err := session.Read(r, sessionKey)
	if err != nil {
		return
	}

	if tokenString != "" {

		var token = new(oauth2.Token)
		err = json.Unmarshal([]byte(tokenString), token)
		if err != nil {
			return
		}

		fmt.Println(token)

		discordDeref.Client = discordOauthConfig.Client(discordOauthContext, token)
	}

	return discordDeref, err
}

//noinspection GoUnusedParameter
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.Bot {
		return
	}

	page, err := websockets.GetPage(websockets.PageChat)
	if err != nil {
		log.Err(err)
		return
	}

	if page.HasConnections() {

		page.Send(chatWebsocketPayload{
			AuthorID:     m.Author.ID,
			AuthorUser:   m.Author.Username,
			AuthorAvatar: m.Author.Avatar,
			Content:      m.Content,
			Channel:      m.ChannelID,
		})
	}
}
