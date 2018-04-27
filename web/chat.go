package web

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/websockets"
)

const (
	guildID          = "407493776597057538"
	generalChannelID = "407493777058693121"
)

var (
	discordSession *discordgo.Session
)

func init() {

	var err error

	// Get client
	discordSession, err = discordgo.New("Bot " + os.Getenv("STEAM_DISCORD_BOT_TOKEN"))
	if err != nil {
		logger.Error(err)
	}

	// Add websocket listener
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if !m.Author.Bot {
			websockets.Send(websockets.CHAT, chatWebsocketPayload{
				AuthorID:     m.Author.ID,
				AuthorUser:   m.Author.Username,
				AuthorAvatar: m.Author.Avatar,
				Content:      m.Content,
			})
		}
	})

	// Open connection
	err = discordSession.Open()
	if err != nil {
		logger.Error(err)
	}
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {

	// Get ID from URL
	id := chi.URLParam(r, "id")
	if id == "" {
		id = generalChannelID
	}

	//
	var wg sync.WaitGroup

	// Get channels
	var channels []*discordgo.Channel
	wg.Add(1)
	go func() {

		channelsResponse, err := discordSession.GuildChannels(guildID)
		if err != nil {
			logger.Error(err)
		}

		for _, v := range channelsResponse {
			if v.Type == discordgo.ChannelTypeGuildText {

				// Fix channel name
				v.Name = strings.Title(strings.Replace(v.Name, "-", " ", 1))

				channels = append(channels, v)
			}
		}

		wg.Done()

	}()

	// Get messages
	var messages []*discordgo.Message
	wg.Add(1)
	go func() {

		messagesResponse, err := discordSession.ChannelMessages(id, 50, "", "", "")
		if err != nil {
			logger.Error(err)
		}

		for _, v := range messagesResponse {
			if !v.Author.Bot {
				messages = append(messages, v)
			}
		}

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	template := chatTemplate{}
	template.Fill(w, r, "Chat")
	template.Channels = channels
	template.Messages = messages
	template.ChannelID = id

	returnTemplate(w, r, "chat", template)
}

type chatTemplate struct {
	GlobalTemplate
	Channels  []*discordgo.Channel
	Messages  []*discordgo.Message
	ChannelID string // Selected channel
}

type chatWebsocketPayload struct {
	AuthorID     string `json:"author_id"`
	AuthorUser   string `json:"author_user"`
	AuthorAvatar string `json:"author_avatar"`
	Content      string `json:"content"`
}
