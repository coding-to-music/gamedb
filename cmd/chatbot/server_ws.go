package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var rateLimit = rate.New(time.Second*3, rate.WithBurst(3))

func websocketServer() (session *discordgo.Session, err error) {

	session, err = discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return nil, err
	}

	session.AddHandler(func(s *discordgo.Session, e *discordgo.InteractionCreate) { interactionHandler(s, e.Interaction) })
	session.AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) { messageHandler(s, e.Message) })
	session.AddHandler(func(s *discordgo.Session, e *discordgo.GuildCreate) { guildHandler(e.Guild) }) // When bot joins a guild

	log.Info("Starting chatbot websocket connection")
	err = session.Open()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func interactionHandler(s *discordgo.Session, i *discordgo.Interaction) {

	// Ignore PMs
	// member is sent when the command is invoked in a guild, and user is sent when invoked in a DM
	// todo, make PR to add user with isDM() func
	// todo, if command.AllowDM() then use user and not member
	if i.Member == nil {
		return
	}

	// Stop users getting two responses
	if config.IsLocal() && i.Member.User.ID != config.DiscordAdminID {
		return
	}

	// Check for pings
	if i.Type == discordgo.InteractionPing {

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		}

		err := s.InteractionRespond(i, response)
		if err != nil {
			log.ErrS(err)
		}
		return
	}

	// Get command
	command, ok := chatbot.CommandCache[i.Data.Name]
	if !ok {
		log.ErrS("Command ID not found in register")
		return
	}

	// Save stats
	var success bool
	defer saveToDB(command, true, &success, argumentsString(i), i.GuildID, i.ChannelID, i.Member.User)

	// Typing notification
	// todo Remove this when slash commands have `thinking`
	err := s.ChannelTyping(i.ChannelID)
	discordError(err)

	//
	code := getProdCC(command, i.Member.User.ID)

	cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), arguments(i), code)

	// Check in cache first
	if !command.DisableCache() && !config.IsLocal() {

		var response = &discordgo.InteractionResponse{}
		err = memcache.GetInterface(cacheItem.Key, &response)
		if err == nil {

			err = s.InteractionRespond(i, response)
			if err != nil {
				log.ErrS(err)
			}
			return
		}
	}

	// Rate limit
	if !rateLimit.GetLimiter(i.Member.User.ID).Allow() {
		log.Warn("over chatbot rate limit", zap.String("author", i.Member.User.ID), zap.String("msg", argumentsString(i)))
		return
	}

	// Make output
	out, err := command.Output(i.Member.User.ID, code, arguments(i))
	if err != nil {
		log.ErrS(err)
		return
	}

	// Convert to slash format
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: out.Content,
		},
	}

	if out.Embed != nil {
		response.Data.Embeds = []*discordgo.MessageEmbed{out.Embed}
	}

	// Respond
	err = s.InteractionRespond(i, response)
	if err != nil {
		log.ErrS(err)
	}

	// Save to cache
	err = memcache.SetInterface(cacheItem.Key, response, cacheItem.Expiration)
	if err != nil {
		log.Err("Saving to memcache", zap.Error(err), zap.String("msg", argumentsString(i)))
	}

	success = true
}

func messageHandler(s *discordgo.Session, m *discordgo.Message) {

	// Don't reply to bots
	if m.Author.Bot {
		return
	}

	// Stop users getting two responses
	if config.IsLocal() && m.Author.ID != config.DiscordAdminID {
		return
	}

	// Scan commands
	for _, command := range chatbot.CommandRegister {

		msg := strings.TrimSpace(m.Content)

		if chatbot.RegexCache[command.Regex()].MatchString(msg) {

			func() { // In a func for the defer

				// Ignore PMs
				private := func() bool {
					channel, err := s.State.Channel(m.ChannelID)
					if err != nil {
						channel, err = s.Channel(m.ChannelID)
						if err != nil {
							discordError(err)
							return false
						}
					}
					return channel.Type == discordgo.ChannelTypeDM
				}()

				if !command.AllowDM() && private {
					return
				}

				// Save stats
				var success bool
				defer saveToDB(command, false, &success, msg, m.GuildID, m.ChannelID, m.Author)

				// Typing notification
				err := s.ChannelTyping(m.ChannelID)
				discordError(err)

				// Get user settings
				code := getProdCC(command, m.Author.ID)

				cacheItem := memcache.ItemChatBotRequest(msg, code)

				// Check in cache first
				if !command.DisableCache() && !config.IsLocal() {
					var message discordgo.MessageSend
					err = memcache.GetInterface(cacheItem.Key, &message)
					if err == nil {
						_, err = s.ChannelMessageSendComplex(m.ChannelID, &message)
						discordError(err)
						return
					}
				}

				// Rate limit
				if !rateLimit.GetLimiter(m.Author.ID).Allow() {
					log.Warn("over chatbot rate limit", zap.String("author", m.Author.ID), zap.String("msg", msg))
					return
				}

				// Make output
				message, err := command.Output(m.Author.ID, code, command.LegacyInputs(msg))
				if err != nil {
					log.WarnS(err, msg)
					return
				}

				// Reply
				_, err = s.ChannelMessageSendComplex(m.ChannelID, &message)
				if err != nil {
					discordError(err)
					return
				}

				// Save to cache
				err = memcache.SetInterface(cacheItem.Key, message, cacheItem.Expiration)
				if err != nil {
					log.ErrS(err, msg)
				}

				success = true
			}()

			break
		}
	}
}

func guildHandler(guild *discordgo.Guild) {

	if guild.MemberCount == 0 {
		return
	}

	mongoGuild := mongo.DiscordGuild{
		ID:      guild.ID,
		Name:    guild.Name,
		Icon:    guild.IconURL(),
		Members: guild.MemberCount,
	}

	ops := options.Update()
	ops.SetUpsert(true)

	_, err := mongo.UpdateOne(mongo.CollectionDiscordGuilds, bson.D{{"_id", guild.ID}}, mongoGuild.BSON(), ops)
	if err != nil {
		log.Err("Updating guild row", zap.Error(err))
	}
}

func getProdCC(command chatbot.Command, authorID string) steamapi.ProductCC {

	code := steamapi.ProductCCUS
	if command.PerProdCode() {
		settings, err := mysql.GetChatBotSettings(authorID)
		if err != nil {
			log.ErrS(err)
		}
		code = settings.ProductCode
	}
	return code
}

func saveToDB(command chatbot.Command, isSlash bool, wasSuccess *bool, message, guildID, channelID string, user *discordgo.User) {

	if user.ID == config.DiscordAdminID {
		return
	}

	if config.IsLocal() {
		return
	}

	if !isSlash && (command.ID() == chatbot.CHelp || command.ID() == chatbot.CInvite) {
		return
	}

	// Influx
	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementChatBot),
		Tags: map[string]string{
			"guild_id":   guildID,
			"channel_id": channelID,
			"author_id":  user.ID,
			"command_id": command.ID(),
			"slash":      strconv.FormatBool(isSlash),
			"success":    strconv.FormatBool(*wasSuccess),
		},
		Fields: map[string]interface{}{
			"request": 1,
		},
		Time:      time.Now(),
		Precision: "ms",
	}

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err)
	}

	var row = mongo.ChatBotCommand{
		GuildID:      guildID,
		ChannelID:    channelID,
		AuthorID:     user.ID,
		AuthorName:   user.Username,
		AuthorAvatar: user.Avatar,
		CommandID:    command.ID(),
		Message:      message,
		Slash:        isSlash,
		Time:         time.Now(), // Can get from ws message?
	}

	_, err = mongo.InsertOne(mongo.CollectionChatBotCommands, row)
	if err != nil {
		log.ErrS(err)
	}

	// Websocket
	guilds, err := mongo.GetGuildsByIDs([]string{row.GuildID})
	if err != nil {
		log.ErrS(err)
	}

	wsPayload := queue.ChatBotPayload{}
	wsPayload.RowData = row.GetTableRowJSON(guilds)

	err = queue.ProduceWebsocket(wsPayload, websockets.PageChatBot)
	if err != nil {
		log.ErrS(err)
	}
}

func arguments(event *discordgo.Interaction) (a map[string]string) {

	a = map[string]string{}
	for _, v := range event.Data.Options {
		a[v.Name] = fmt.Sprint(v.Value)
	}
	return a
}

func argumentsString(event *discordgo.Interaction) string {

	var s = []string{event.Data.Name}
	for _, v := range event.Data.Options {
		s = append(s, fmt.Sprint(v.Value))
	}
	return strings.Join(s, " ")
}

func discordError(err error) {

	var allowed = map[int]string{
		50001: "Missing Access",
		50013: "Missing Permissions",
	}

	if err != nil {
		if val, ok := err.(*discordgo.RESTError); ok {
			if _, ok2 := allowed[val.Message.Code]; ok2 {
				zap.S().Info(err) // No helper to fix stack offset
				return
			}
		}

		zap.S().Error(err) // No helper to fix stack offset
		return
	}
}
