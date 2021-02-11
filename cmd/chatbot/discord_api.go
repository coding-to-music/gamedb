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
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func refreshCommands() error {

	apiCommands, err := getCommands()
	if err != nil {
		return err
	}

	// Delete removed commands
	for _, apiCommand := range apiCommands {
		if _, ok := chatbot.CommandCache[apiCommand.Name]; !ok {

			log.Info("Deleting dommand", zap.String("id", apiCommand.Name))
			code, err := deleteCommand(apiCommand.ID)
			if err != nil {
				log.Err("Deleting command", zap.Int("code", code), zap.Error(err))
			}
		}
	}

	// Update updated commands
	for _, apiCommand := range apiCommands {
		if localCommand, ok := chatbot.CommandCache[apiCommand.Name]; ok {

			if apiCommand.Options == nil {
				apiCommand.Options = []interactions.InteractionOption{}
			}

			b1, _ := json.Marshal(apiCommand.Options)
			b2, _ := json.Marshal(localCommand.Slash())
			if string(b1) != string(b2) {

				log.Info("Updating command", zap.String("id", localCommand.ID()))
				err = upsertCommand(localCommand)
				if err != nil {
					return err
				}
			}
		}
	}

	// Add missing commands
	for k, localCommand := range chatbot.CommandCache {
		func() {

			// Check if already exists
			for _, apiCommand := range apiCommands {
				if apiCommand.Name == k {
					return
				}
			}

			log.Info("Adding command", zap.String("id", localCommand.ID()))
			err = upsertCommand(localCommand)
			if err != nil {
				log.Err("Adding command", zap.String("id", localCommand.ID()))
				return
			}
		}()
	}

	return nil
}

func getCommands() (ints []interactions.Interaction, err error) {

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	b, _, err := helpers.Get("https://discord.com/api/v8/applications/"+config.DiscordBotClientID+"/commands", 0, headers)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &ints)

	return ints, err
}

func upsertCommand(command chatbot.Command) error {

	time.Sleep(time.Second)

	payload := interactions.Interaction{
		Name:        command.ID(),
		Description: strings.ToUpper(string(command.Type())) + ": " + command.Description(),
		Options:     command.Slash(),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	path := "https://discord.com/api/v8/applications/" + config.DiscordBotClientID + "/commands"
	req, err := http.NewRequest("POST", path, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header = http.Header{}
	req.Header.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	req.Header.Set("Content-Type", "application/json")

	clientWithTimeout := &http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := clientWithTimeout.Do(req)
	if err != nil {
		return err
	}

	defer helpers.Close(resp.Body)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Err("Upserting discord command", zap.Int("code", resp.StatusCode), zap.String("id", command.ID()), zap.String("body", string(body)))
	}

	return nil
}

func deleteCommand(id string) (int, error) {

	time.Sleep(time.Second)

	headers := http.Header{}
	headers.Set("Authorization", "Bot "+config.C.DiscordChatBotToken)
	headers.Set("Content-Type", "application/json")

	_, code, err := helpers.Delete("https://discord.com/api/v8/applications/"+config.DiscordBotClientID+"/commands/"+id, 0, headers)
	return code, err
}
