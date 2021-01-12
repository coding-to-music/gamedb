package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/discord"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

//goland:noinspection GoUnusedFunction
func refreshCommands() error {

	ints, err := getCommands()
	if err != nil {
		return err
	}

	for _, v := range ints {
		if _, ok := chatbot.CommandCache[v.Name]; !ok {

			log.Info("Deleting " + v.Name)

			err := deleteCommand(v.ID)
			if err != nil {
				log.Err("deleting old discord command", zap.Error(err))
			}
		}
	}

	setCommands()
	return nil
}

//goland:noinspection GoUnusedFunction
func deleteCommand(id string) error {

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	_, _, err := helpers.Delete("https://discord.com/api/v8/applications/"+discord.ClientIDBot+"/commands/"+id, 0, headers)
	return err
}

//goland:noinspection GoUnusedFunction
func getCommands() (ints []interactions.Interaction, err error) {

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	b, _, err := helpers.Get("https://discord.com/api/v8/applications/"+discord.ClientIDBot+"/commands", 0, headers)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &ints)

	return ints, err
}

//goland:noinspection GoUnusedFunction
func setCommands() {

	path := "https://discord.com/api/v8/applications/" + discord.ClientIDBot + "/commands"

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	for _, c := range chatbot.CommandRegister {

		func(c chatbot.Command) {

			payload := interactions.Interaction{
				Name:        c.ID(),
				Description: strings.ToUpper(string(c.Type())) + ":" + c.Description(),
				Options:     c.Slash(),
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

			defer helpers.Close(resp.Body)

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.ErrS(err)
				return
			}

			log.Info("Command updated", zap.Int("code", resp.StatusCode), zap.String("id", c.ID()))

			if resp.StatusCode != 200 && resp.StatusCode != 201 {
				log.Err("Upserting discord command", zap.Int("code", resp.StatusCode), zap.String("id", c.ID()), zap.String("body", string(body)))
			}

			time.Sleep(time.Second / 4)
		}(c)
	}
}
