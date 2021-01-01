package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
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

	if !config.IsLocal() {
		return nil
	}

	path := "https://discord.com/api/v8/applications/" + discord.ClientIDBot + "/commands"

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	for _, c := range chatbot.CommandRegister {

		if val, ok := c.(chatbot.SlashCommandInterface); ok {

			b, err := json.Marshal(val.Slash())
			if err != nil {
				return err
			}

			req, err := http.NewRequest("POST", path, bytes.NewBuffer(b))
			if err != nil {
				return err
			}

			req.Header = headers

			clientWithTimeout := &http.Client{
				Timeout: time.Second * 2,
			}

			resp, err := clientWithTimeout.Do(req)
			if err != nil {
				return err
			}

			//goland:noinspection GoDeferInLoop
			defer helpers.Close(resp.Body)

			// body, err := ioutil.ReadAll(resp.Body)
			// if err != nil {
			// 	return err
			// }
			//
			// log.Info(string(body), zap.Int("code", resp.StatusCode))
		}
	}

	// Get all
	b, _, err := helpers.Get(path, 0, headers)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = json.Indent(buf, b, "", "  ")
	if err != nil {
		return err
	}

	log.Info(buf.String())

	return nil
}
