package chat_bot

import (
	"regexp"
	"strings"

	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/mongo"
)

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
	return "Get info on a player"
}

func (CommandGetPlayer) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(player|user) (.*)")
}

func (c CommandGetPlayer) Output(input string) string {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[2])
	if err != nil {
		log.Err(err)
		return ""
	}

	return player.GetName()
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
