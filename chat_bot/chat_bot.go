package chat_bot

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/config"
)

func Init() {

	discord, err := discordgo.New("Bot " + config.Config.DiscordBotToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		if config.Config.IsLocal() && m.Author.ID != "145456943912189952" {
			return
		}

		if m.Author.Bot {
			return
		}

		for _, v := range CommandRegister {

			if v.Regex().MatchString(m.Message.Content) {

				private, err := isPrivate(s, m)
				if err != nil {
					fmt.Println(err)
					return
				}

				chanID := m.ChannelID

				if private {

					st, err := s.UserChannelCreate(m.Author.ID)
					if err != nil {
						fmt.Println(err)
						return
					}

					chanID = st.ID
				}

				_, err = s.ChannelMessageSend(chanID, v.Output(m.Message.Content))
				if err != nil {
					fmt.Println(err)
					return
				}

				return
			}
		}
	})

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func isPrivate(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}

var CommandRegister = []Command{
	CommandGetPlayer{},
	CommandHelp{},
	CommandAppPlayers{},
}

type Command interface {
	Description() string
	Regex() *regexp.Regexp
	Output(input string) string
	Example() string
}

//
type CommandGetPlayer struct {
}

func (CommandGetPlayer) Description() string {
	return "Links to a list of commands"
}

func (CommandGetPlayer) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.player [a-zA-Z0-9]+")
}

func (CommandGetPlayer) Output(input string) string {
	return "" // todo
}

func (CommandGetPlayer) Example() string {
	return ".player {playerName}"
}

//
type CommandAppPlayers struct {
}

func (CommandAppPlayers) Description() string {
	return "Gets the number of people playing."
}

func (CommandAppPlayers) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.players [a-zA-Z0-9]+")
}

func (CommandAppPlayers) Output(input string) string {

	input = strings.TrimPrefix(input, ".players ")

	return "" // todo
}

func (CommandAppPlayers) Example() string {
	return ".player {playerName}"
}

//
type CommandHelp struct {
}

func (CommandHelp) Description() string {
	return "Links to a players profile"
}

func (CommandHelp) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.help")
}

func (CommandHelp) Output(input string) string {
	return "http://gamedb.online/chat-bot"
}

func (CommandHelp) Example() string {
	return ".player {playerName}"
}

// .game 123 |.app half life
// .user 123 |.user jimeagle
// .recent 123|jimeagle
// .trending - top 10
// .popular - top 10 based on players
