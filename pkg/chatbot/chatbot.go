package chatbot

import (
	"regexp"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type CommandType string

var (
	TypeGame   CommandType = "Game"
	TypePlayer CommandType = "Player"
	TypeGroup  CommandType = "Group"
	TypeOther  CommandType = "Miscellaneous"

	RegexCache   = make(map[string]*regexp.Regexp, len(CommandRegister))
	CommandCache = make(map[string]Command, len(CommandRegister))
)

type Command interface {
	ID() string
	Regex() string
	DisableCache() bool
	PerProdCode() bool
	Output(authorID string, region steamapi.ProductCC, inputs map[string]string) (discordgo.MessageSend, error)
	Example() string
	Description() string
	Type() CommandType
	Slash() []interactions.InteractionOption
	LegacyInputs(input string) map[string]string
}

// These are the discord slash command names, if changed, the old one needs to be deleted
const (
	CApp            = "game"            //
	CAppFollowers   = "followers"       //
	CAppPlayers     = "players"         //
	CAppPrice       = "price"           //
	CAppsRandom     = "random"          //
	CAppsNew        = "new"             //
	CAppsPopular    = "top"             //
	CAppsTrending   = "trending-games"  //
	CGroup          = "group"           //
	CGroupsTrending = "trending-groups" //
	CPlayer         = "player"          //
	CPlayerApps     = "games"           // Count
	CPlayerLevel    = "level"           //
	CPlayerPlaytime = "playtime"        //
	CPlayerRecent   = "recent"          //
	CPlayerUpdate   = "update"          //
	CPlayerWishlist = "wishlist"        //
	CPlayerLibrary  = "library"         //
	CHelp           = "help"            //
	CInvite         = "invite"          //
	CSettings       = "settings"        //
	CSteamOnline    = "online"          //
)

var CommandRegister = []Command{
	&CommandApp{},
	&CommandAppFollowers{},
	&CommandAppPlayers{},
	&CommandSteamOnline{},
	&CommandAppRandom{},
	&CommandAppsNew{},
	&CommandAppPrice{},
	&CommandAppsPopular{},
	&CommandAppsTrending{},
	&CommandGroup{},
	&CommandGroupsTrending{},
	&CommandPlayer{},
	&CommandPlayerApps{},
	&CommandPlayerLevel{},
	&CommandPlayerPlaytime{},
	&CommandPlayerRecent{},
	&CommandPlayerLibrary{},
	&CommandPlayerUpdate{},
	&CommandPlayerWishlist{},
	&CommandHelp{},
	&CommandInvite{},
	&CommandSettings{},
}

func init() {
	for _, v := range CommandRegister {
		RegexCache[v.Regex()] = regexp.MustCompile(v.Regex())
		CommandCache[v.ID()] = v
	}
}

func getAuthor(guildID string) *discordgo.MessageEmbedAuthor {

	author := &discordgo.MessageEmbedAuthor{
		Name:    "gamedb.online",
		URL:     config.C.GameDBDomain + "/?utm_source=discord&utm_medium=discord&utm_content=" + guildID,
		IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
	}
	if config.IsLocal() {
		author.Name = "localhost:" + config.C.FrontendPort
		author.URL = "http://localhost:" + config.C.FrontendPort + "/"
	}
	return author
}

func getFooter() *discordgo.MessageEmbedFooter {

	footer := &discordgo.MessageEmbedFooter{
		Text:         "gamedb.online/discord for all commands",
		IconURL:      "https://gamedb.online/assets/img/sa-bg-32x32.png",
		ProxyIconURL: "",
	}

	if config.IsLocal() {
		footer.Text = "LOCAL"
	}

	return footer
}

type App interface {
	GetID() int
	GetName() string
	GetPath() string
	GetHeaderImage() string
	GetPlayersPeakWeek() int
	GetFollowers() string
	GetPrices() helpers.ProductPrices
	GetReviewScore() string
	GetReleaseDateNice() string
}

func getAppEmbed(commandID string, app App, code steamapi.ProductCC) *discordgo.MessageEmbed {

	return &discordgo.MessageEmbed{
		Title:     app.GetName(),
		URL:       config.C.GameDBDomain + app.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage()},
		Footer:    getFooter(),
		Color:     2664261,
		Image: &discordgo.MessageEmbedImage{
			URL: charts.GetAppPlayersChart(commandID, app.GetID(), "168d", "1d"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Max Weekly Players",
				Value:  humanize.Comma(int64(app.GetPlayersPeakWeek())),
				Inline: true,
			},
			{
				Name:   "Followers",
				Value:  app.GetFollowers(),
				Inline: true,
			},
			{
				Name:   "\u200B",
				Value:  "\u200B",
				Inline: true,
			},
			{
				Name:   "Price",
				Value:  app.GetPrices().Get(code).GetFinal(),
				Inline: true,
			},
			{
				Name:   "Review Score",
				Value:  app.GetReviewScore(),
				Inline: true,
			},
			{
				Name:   "\u200B",
				Value:  "\u200B",
				Inline: true,
			},
			{
				Name:   "Release Date",
				Value:  app.GetReleaseDateNice(),
				Inline: true,
			},
		},
	}
}
