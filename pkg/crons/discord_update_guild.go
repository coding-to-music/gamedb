package crons

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type DiscordUpdateGuild struct {
	BaseTask
}

func (c DiscordUpdateGuild) ID() string {
	return "discord-update-guild"
}

func (c DiscordUpdateGuild) Name() string {
	return "Update the most stale discord guild"
}

func (c DiscordUpdateGuild) Group() TaskGroup {
	return ""
}

func (c DiscordUpdateGuild) Cron() TaskTime {
	return CronTimeUpdateDiscordGuild
}

func (c DiscordUpdateGuild) work() (err error) {

	guilds, err := mongo.GetGuilds(0, 1, bson.D{{"update_at", 1}}, nil)
	if err != nil {
		return err
	}
	if len(guilds) == 0 {
		return errors.New("no results")
	}

	discord, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return err
	}

	for _, guild := range guilds {

		resp, err := discord.GuildPreview(guild.ID) // Swap to Guild() when PR is in
		if err != nil {

			if val, ok := err.(*discordgo.RESTError); ok {
				if val.Response.StatusCode == 404 {
					_, err = mongo.ReplaceOne(mongo.CollectionDiscordGuilds, bson.D{{"_id", guild.ID}}, guild)
					if err != nil {
						return err
					}
					continue
				}
			}

			return err
		}

		if resp.ApproximateMemberCount == 0 {
			continue
		}

		fullGuild := discordgo.Guild{Icon: resp.Icon} // todo, remove when pr gets merged in

		guild.ID = resp.ID
		guild.Name = resp.Name
		guild.Icon = fullGuild.Icon
		guild.Members = resp.ApproximateMemberCount

		_, err = mongo.ReplaceOne(mongo.CollectionDiscordGuilds, bson.D{{"_id", resp.ID}}, guild)
		if err != nil {
			return err
		}
	}

	return nil
}
