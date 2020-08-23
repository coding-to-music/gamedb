package chatbot

import (
	"html/template"
	"regexp"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
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
	ID() string
	Regex() string
	DisableCache() bool
	Output(*discordgo.MessageCreate) (discordgo.MessageSend, error)
	Example() string
	Description() template.HTML
	Type() CommandType
}

const (
	CApp            = "app"
	CSettings       = "settings"
	CAppPlayers     = "app-players"
	CAppRandom      = "app-random"
	CAppPrice       = "app-price"
	CAppsNew        = "apps-new"
	CAppsPopular    = "apps-popular"
	CAppsTrending   = "apps-trending"
	CGroup          = "group"
	CGroupsTrending = "groups-trending"
	CPlayer         = "player"
	CPlayerApps     = "player-apps"
	CPlayerLevel    = "player-level"
	CPlayerPlaytime = "player-playtime"
	CPlayerRecent   = "player-recent"
	CPlayerUpdate   = "player-update"
	CPlayerWishlist = "player-wishlist"
	CPlayerLibrary  = "player-library"
	CHelp           = "help"
	CSteamOnline    = "steam-online"
)

var CommandRegister = []Command{
	CommandApp{},
	CommandAppPlayers{},
	CommandSteamOnline{},
	CommandAppRandom{},
	CommandAppsNew{},
	CommandAppPrice{},
	CommandAppsPopular{},
	CommandAppsTrending{},
	CommandGroup{},
	CommandGroupsTrending{},
	CommandPlayer{},
	CommandPlayerApps{},
	CommandPlayerLevel{},
	CommandPlayerPlaytime{},
	CommandPlayerRecent{},
	CommandPlayerLibrary{},
	CommandPlayerUpdate{},
	CommandPlayerWishlist{},
	CommandHelp{},
	CommandSettings{},
}

func init() {
	for _, v := range CommandRegister {
		RegexCache[v.Regex()] = regexp.MustCompile(v.Regex())
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
		Text:         "Powered by gamedb.online",
		IconURL:      "https://gamedb.online/assets/img/sa-bg-32x32.png",
		ProxyIconURL: "",
	}

	if config.IsLocal() {
		footer.Text = "LOCAL"
	}

	return footer
}

func getAppEmbed(app mongo.App) *discordgo.MessageEmbed {

	zap.S().Info(charts.GetAppChart(app))

	return &discordgo.MessageEmbed{
		Title:     app.GetName(),
		URL:       config.C.GameDBDomain + app.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage()},
		Footer:    getFooter(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Max Weekly Players",
				Value: humanize.Comma(int64(app.PlayerPeakWeek)),
			},
			{
				Name:  "Release Date",
				Value: app.GetReleaseDateNice(),
			},
			{
				Name:  "Price",
				Value: app.Prices.Get(steamapi.ProductCCUS).GetFinal(),
			},
			{
				Name:  "Review Score",
				Value: app.GetReviewScore(),
			},
			{
				Name:  "Followers",
				Value: app.GetFollowers(),
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: charts.GetAppChart(app),
		},
	}
}
