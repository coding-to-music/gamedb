package web

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/spf13/viper"
)

const (
	guildID          = "407493776597057538"
	generalChannelID = "407493777058693121"
)

var (
	discordSession *discordgo.Session
)

// Called from main
func InitChat() {

	// Retry connecting to Discord
	for i := 1; i <= 3; i++ {
		err := connectToDiscord()
		if err != nil && i < 3 {
			time.Sleep(time.Second)
		} else if err != nil {
			log.Log(err)
		} else {
			break
		}
	}
}

func chatRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", ChatHandler)
	r.Get("/{id}", ChatHandler)
	return r
}

func connectToDiscord() error {

	var err error

	// Get client
	discordSession, err = discordgo.New("Bot " + viper.GetString("DISCORD_BOT_TOKEN"))
	if err != nil {
		return err
	}

	// Add websocket listener
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !m.Author.Bot {

			page, err := websockets.GetPage(websockets.PageChat)
			if err == nil && page.HasConnections() {

				page.Send(chatWebsocketPayload{
					AuthorID:     m.Author.ID,
					AuthorUser:   m.Author.Username,
					AuthorAvatar: m.Author.Avatar,
					Content:      m.Content,
				})
			}
		}
	})

	// Open connection
	return discordSession.Open()
}

type chatWebsocketPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorUser   string `json:"author_user"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := chatTemplate{}
	t.Fill(w, r, "Chat")

	setNoCacheHeaders(w)

	// Get ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		id = generalChannelID
	}
	t.ChannelID = id

	//
	var wg sync.WaitGroup

	// Get channels
	wg.Add(1)
	go func() {

		defer wg.Done()

		channelsResponse, err := discordSession.GuildChannels(guildID)
		log.Log(err)

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

	// Get messages
	wg.Add(1)
	go func() {

		defer wg.Done()

		messagesResponse, err := discordSession.ChannelMessages(id, 50, "", "", "")
		log.Log(err)

		for _, v := range messagesResponse {
			if !v.Author.Bot && v.Type == discordgo.MessageTypeDefault {
				t.Messages = append(t.Messages, v)
			}
		}

	}()

	// Get members
	wg.Add(1)
	go func() {

		defer wg.Done()

		membersResponse, err := discordSession.GuildMembers(guildID, "", 1000)
		log.Log(err)

		for _, v := range membersResponse {
			if !v.User.Bot {
				t.Members = append(t.Members, v)
			}
		}

	}()

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "chat", t)
	log.Log(err)
}

type chatTemplate struct {
	GlobalTemplate
	ChannelID string
	Channels  []*discordgo.Channel
	Messages  []*discordgo.Message
	Members   []*discordgo.Member
}
