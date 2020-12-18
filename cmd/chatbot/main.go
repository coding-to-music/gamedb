package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"time"

	"github.com/gamedb/gamedb/pkg/chatbot"
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

	err = slashCommandRegister()
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

func slashCommandRegister() error {

	for _, c := range chatbot.CommandRegister {

		if c.ID() != chatbot.CAppPlayers || !config.IsLocal() {
			continue
		}

		vals := url.Values{}

		path := "https://discord.com/api/v8/applications/" + discord.ClientIDBot + "/commands/" + c.ID()

		req, err := http.NewRequest("POST", path, bytes.NewBufferString(vals.Encode()))
		if err != nil {
			return err
		}

		req.Header = http.Header{}
		req.Header.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)

		clientWithTimeout := &http.Client{
			Timeout: time.Second * 2,
		}

		resp, err := clientWithTimeout.Do(req)
		if err != nil {
			return err
		}

		//goland:noinspection GoDeferInLoop
		defer helpers.Close(resp.Body)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		log.Info(string(body), zap.Int("code", resp.StatusCode))
	}

	return nil
}

type SlashCommand struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Options     []SlashCommandOption `json:"options"`
}

type SlashCommandOption struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Type        int                        `json:"type"`
	Required    bool                       `json:"required"`
	Choices     []SlashCommandOptionChoice `json:"choices,omitempty"`
}

type SlashCommandOptionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
