package discord

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	sessions = map[string]*discordgo.Session{}
	lock     sync.Mutex
)

func GetDiscordBot(authToken string, bot bool, handlers ...interface{}) (session *discordgo.Session, err error) {

	lock.Lock()
	defer lock.Unlock()

	if bot {
		authToken = "Bot " + authToken
	} else {
		authToken = "Bearer " + authToken
	}

	_, ok := sessions[authToken]
	if !ok {
		discord, err := discordgo.New(authToken)
		if err != nil {
			return discord, err
		}

		if bot {
			for _, v := range handlers {
				discord.AddHandler(v)
			}

			// Open connection
			err = discord.Open()
			if err != nil {
				return discord, err
			}
		}

		sessions[authToken] = discord
	}

	return sessions[authToken], err
}
