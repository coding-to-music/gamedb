package chatbot

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type CommandType string

var (
	TypeGame   CommandType = "Game"
	TypePlayer CommandType = "Player"
	TypeGroup  CommandType = "Group"
	TypeOther  CommandType = "Miscellaneous"

	RegexCache = make(map[string]*regexp.Regexp, len(CommandRegister))
)

type Command interface {
	Regex() string
	DisableCache() bool
	Output(*discordgo.MessageCreate) (discordgo.MessageSend, error)
	Example() string
	Description() string
	Type() CommandType
}

var CommandRegister = []Command{
	CommandApp{},
	CommandAppPlayers{},
	CommandAppPlayersSteam{},
	CommandAppRandom{},
	CommandAppsNew{},
	CommandAppsPopular{},
	CommandAppsTrending{},
	CommandGroup{},
	CommandGroupsTrending{},
	CommandPlayer{},
	CommandPlayerApps{},
	CommandPlayerLevel{},
	CommandPlayerPlaytime{},
	CommandPlayerRecent{},
	CommandHelp{},
}

func getAuthor(guildID string) *discordgo.MessageEmbedAuthor {

	author := &discordgo.MessageEmbedAuthor{
		Name:    "gamedb.online",
		URL:     "https://gamedb.online/?utm_source=discord&utm_medium=discord&utm_content=" + guildID,
		IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
	}
	if config.IsLocal() {
		author.Name = "localhost:" + config.Config.WebserverPort.Get()
		author.URL = "http://localhost:" + config.Config.WebserverPort.Get() + "/"
	}
	return author
}

func getFooter() *discordgo.MessageEmbedFooter {

	footer := &discordgo.MessageEmbedFooter{
		Text:         "gamedb.online",
		IconURL:      "https://gamedb.online/assets/img/sa-bg-32x32.png",
		ProxyIconURL: "",
	}

	if config.IsLocal() {
		footer.Text = "localhost:" + config.Config.WebserverPort.Get()
	}

	return footer
}

func saveChartToFile(b []byte, filename string) error {

	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		log.Err(err)
	}()

	_, err = f.Write(b)
	return err
}

func saveChartToGoogle(b []byte, filename string) (string, error) {

	client, ctx, err := helpers.GetStorageClient()
	if err != nil {
		return "", err
	}

	w := client.Bucket(helpers.BucketChatBot).Object(filename).NewWriter(ctx)

	_, err = io.Copy(w, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	opts, err := helpers.GetSignedURLOptions()
	if err != nil {
		return "", err
	}

	return storage.SignedURL(helpers.BucketChatBot, filename, opts)
}
