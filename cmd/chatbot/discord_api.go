package main

import (
	"encoding/json"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func refreshCommands(session *discordgo.Session) error {

	apiCommands, err := session.ApplicationCommands("", "")
	if err != nil {
		return err
	}

	// Delete removed commands
	for _, apiCommand := range apiCommands {
		if _, ok := chatbot.CommandCache[apiCommand.Name]; !ok {

			log.Info("Deleting dommand", zap.String("id", apiCommand.Name))
			err = session.ApplicationCommandDelete("", "", apiCommand.ID)
			if err != nil {
				log.Err("Deleting command", zap.Error(err))
			}
		}
	}

	// Update updated commands
	for _, apiCommand := range apiCommands {
		if localCommand, ok := chatbot.CommandCache[apiCommand.Name]; ok {

			if apiCommand.Options == nil {
				apiCommand.Options = []*discordgo.ApplicationCommandOption{}
			}

			b1, _ := json.Marshal(apiCommand.Options)
			b2, _ := json.Marshal(localCommand.Slash())
			if string(b1) != string(b2) {

				log.Info("Updating command", zap.String("id", localCommand.ID()))
				command := &discordgo.ApplicationCommand{
					Name:        localCommand.ID(),
					Description: strings.ToUpper(string(localCommand.Type())) + ": " + localCommand.Description(),
					Options:     localCommand.Slash(),
				}
				_, err = session.ApplicationCommandCreate("", "", command)
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
			command := &discordgo.ApplicationCommand{
				Name:        localCommand.ID(),
				Description: strings.ToUpper(string(localCommand.Type())) + ": " + localCommand.Description(),
				Options:     localCommand.Slash(),
			}
			_, err = session.ApplicationCommandCreate("", "", command)
			if err != nil {
				log.Err("Adding command", zap.String("id", localCommand.ID()))
				return
			}
		}()
	}

	return nil
}
