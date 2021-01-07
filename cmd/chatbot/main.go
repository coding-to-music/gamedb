package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/discord"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.uber.org/zap"
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameChatbot)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if !config.IsConsumer() {
		go func() {
			err := http.ListenAndServe(":6061", nil)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	err = mysql.GetConsumer("chatbot")
	if err != nil {
		log.ErrS(err)
		return
	}

	if config.IsConsumer() {
		log.Err("Prod & local only")
		return
	}

	queue.Init(queue.ChatbotDefinitions)

	discordSession, err := websocketServer()
	if err != nil {
		log.FatalS(err)
	}

	err = slashCommandServer()
	if err != nil {
		log.FatalS(err)
	}

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		func() {
			_ = discordSession.Close()
		},
		func() {
			influxHelper.GetWriter().Flush()
		},
	)
}

//goland:noinspection GoUnusedFunction
func deleteCommand(id string) {

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	_, code, err := helpers.Delete("https://discord.com/api/v8/applications/"+discord.ClientIDBot+"/commands/"+id, 0, headers)
	log.InfoS(code, err)
}

//goland:noinspection GoUnusedFunction
func getCommands() {

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	b, _, err := helpers.Get("https://discord.com/api/v8/applications/"+discord.ClientIDBot+"/commands", 0, headers)
	if err != nil {
		log.ErrS(err)
		return
	}

	buf := bytes.NewBuffer(nil)
	err = json.Indent(buf, b, "", "  ")
	if err != nil {
		log.ErrS(err)
		return
	}

	log.Info(buf.String())
}

//goland:noinspection GoUnusedFunction
func setCommands() {

	path := "https://discord.com/api/v8/applications/" + discord.ClientIDBot + "/commands"

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	for _, c := range chatbot.CommandRegister {

		if val, ok := c.(chatbot.SlashCommand); ok {

			payload := interactions.Interaction{
				Name:        c.ID(),
				Description: string(c.Description()),
				Options:     val.Slash(),
			}

			b, err := json.Marshal(payload)
			if err != nil {
				log.ErrS(err)
				return
			}

			req, err := http.NewRequest("POST", path, bytes.NewBuffer(b))
			if err != nil {
				log.ErrS(err)
				return
			}

			req.Header = headers

			clientWithTimeout := &http.Client{
				Timeout: time.Second * 2,
			}

			resp, err := clientWithTimeout.Do(req)
			if err != nil {
				log.ErrS(err)
				return
			}

			//goland:noinspection GoDeferInLoop
			defer helpers.Close(resp.Body)

			log.Info("Command updated", zap.Int("code", resp.StatusCode), zap.String("id", c.ID()))
		}
	}
}
