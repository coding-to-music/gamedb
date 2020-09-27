package pages

import (
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"github.com/russross/blackfriday"
)

const (
	generalChannelID = "407493777058693121"
)

var discordRelayBotSession *discordgo.Session

func init() {

	var err error

	discordRelayBotSession, err = discordgo.New("Bot " + config.C.DiscordRelayBotToken)
	if err != nil {
		log.ErrS(err)
		return
	}

	// discordRelayBotSession.AddHandler(func(session *discordgo.Session, message *discordgo.MessageCreate) {
	//
	// 	page := websockets.GetPage(websockets.PageChat)
	// 	if page != nil {
	// 		page.Send(websockets.ChatPayload{
	// 			AuthorID:     message.Author.ID,
	// 			AuthorUser:   message.Author.Username,
	// 			AuthorAvatar: message.Author.Avatar,
	// 			Content:      string(blackfriday.Run([]byte(message.Content), blackfriday.WithNoExtensions())),
	// 			Channel:      message.ChannelID,
	// 			Time:         string(message.Timestamp),
	// 			Embeds:       len(message.Embeds) > 0,
	// 			I:            0,
	// 		})
	// 	}
	// })
	//
	// // Open connection
	// err = discordRelayBotSession.Open()
	// if err != nil {
	// 	log.ErrS(err)
	// }
}

func ChatRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", chatHandler)
	r.Get("/{id}", chatHandler)
	r.Get("/{id}/chat.json", chatAjaxHandler)
	return r
}

func chatHandler(w http.ResponseWriter, r *http.Request) {

	http.Redirect(w, r, "https://discord.gg/c5zrcus", http.StatusFound)

	// // Get ID from URL
	// id := chi.URLParam(r, "id")
	// if id == "" {
	// 	http.Redirect(w, r, "/chat/"+generalChannelID, http.StatusFound)
	// 	return
	// }
	//
	// // Template
	// t := chatTemplate{}
	// t.fill(w, r, "Chat", "The Game DB community.")
	// t.ChannelID = id
	// t.addAssetJSON2HTML()
	//
	// //
	// var wg sync.WaitGroup
	// var discordErr error
	//
	// // Get channels
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var channelsResponse []*discordgo.Channel
	//
	// 	operation := func() (err error) {
	// 		channelsResponse, err = discordRelayBotSession.GuildChannels(guildID)
	// 		return err
	// 	}
	//
	// 	policy := backoff.NewExponentialBackOff()
	//
	// 	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err) })
	// 	if err != nil {
	// 		discordErr = err
	// 		log.FatalS(err)
	// 	}
	//
	// 	for _, v := range channelsResponse {
	// 		if v.Type == discordgo.ChannelTypeGuildText {
	//
	// 			// Fix channel name
	// 			v.Name = strings.Replace(v.Name, "-", " ", 1)
	// 			v.Name = strings.Replace(v.Name, "db", "DB", 1)
	// 			v.Name = strings.Title(v.Name)
	//
	// 			t.Channels = append(t.Channels, v)
	// 		}
	// 	}
	//
	// }()
	//
	// // Get members
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var membersResponse []*discordgo.Member
	//
	// 	operation := func() (err error) {
	// 		membersResponse, err = discordRelayBotSession.GuildMembers(guildID, "", 1000)
	// 		return err
	// 	}
	//
	// 	policy := backoff.NewExponentialBackOff()
	//
	// 	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err) })
	// 	if err != nil {
	// 		discordErr = err
	// 		log.FatalS(err)
	// 	}
	//
	// 	for _, v := range membersResponse {
	// 		if !v.User.Bot {
	// 			t.Members = append(t.Members, v)
	// 		}
	// 	}
	//
	// }()
	//
	// // Wait
	// wg.Wait()
	//
	// if discordErr != nil {
	// 	returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Could not connect to Discord."})
	// 	return
	// }
	//
	// returnTemplate(w, r, "chat", t)
}

// type chatTemplate struct {
// 	globalTemplate
// 	ChannelID string
// 	Channels  []*discordgo.Channel
// 	Members   []*discordgo.Member
// }

func chatAjaxHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		id = generalChannelID
	}

	var messagesResponse []*discordgo.Message

	operation := func() (err error) {
		messagesResponse, err = discordRelayBotSession.ChannelMessages(id, 50, "", "", "")
		return err
	}

	policy := backoff.NewExponentialBackOff()

	err := backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err) })
	if err != nil {
		log.ErrS(err)
		return
	}

	var messages []queue.ChatPayload
	var i float32
	for _, message := range messagesResponse {

		if message.Type == discordgo.MessageTypeDefault {

			messages = append(messages, queue.ChatPayload{
				AuthorID:     message.Author.ID,
				AuthorUser:   message.Author.Username,
				AuthorAvatar: message.Author.Avatar,
				Content:      string(blackfriday.Run([]byte(message.Content), blackfriday.WithNoExtensions())),
				Channel:      message.ChannelID,
				Time:         string(message.Timestamp),
				Embeds:       len(message.Embeds) > 0,
				I:            i / 20,
			})

			i++
		}
	}

	returnJSON(w, r, messages)
}
