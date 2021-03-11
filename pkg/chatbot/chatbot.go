package chatbot

import (
	"regexp"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.uber.org/zap"
)

const greenHexDec = 2664261

type CommandType string

func (ct CommandType) Order() int {
	switch ct {
	case TypeGame:
		return 0
	case TypePlayer:
		return 1
	case TypeGroup:
		return 2
	case TypeOther:
		return 3
	default:
		return 4
	}
}

const (
	TypeGame   CommandType = "Game"
	TypePlayer CommandType = "Player"
	TypeGroup  CommandType = "Group"
	TypeOther  CommandType = "Misc"
)

var (
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
	Slash() []*discordgo.ApplicationCommandOption
	LegacyInputs(input string) map[string]string
	AllowDM() bool
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
	CFeedback       = "feedback"        //
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
	&CommandFeedback{},
}

func init() {
	for _, v := range CommandRegister {
		RegexCache[v.Regex()] = regexp.MustCompile(v.Regex())
		CommandCache[v.ID()] = v
	}
}

func getAuthor(guildID string) *discordgo.MessageEmbedAuthor {

	author := &discordgo.MessageEmbedAuthor{
		Name:    "globalsteam.online",
		URL:     config.C.GlobalSteamDomain + "/?utm_source=discord&utm_medium=discord&utm_content=" + guildID,
		IconURL: config.C.GlobalSteamDomain + "/assets/img/sa-bg-32x32.png",
	}
	if config.IsLocal() {
		author.Name = "localhost:" + config.C.FrontendPort
		author.URL = "http://localhost:" + config.C.FrontendPort + "/"
	}
	return author
}

func getFooter() *discordgo.MessageEmbedFooter {

	footer := &discordgo.MessageEmbedFooter{
		Text:         "globalsteam.online/discord for all commands",
		IconURL:      "https://globalsteam.online/assets/img/sa-bg-32x32.png", // Use domain to hotlink
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
	GetPathAbsolute() string
	GetHeaderImage() string
	GetPlayersPeakWeek() int
	GetFollowers() string
	GetPrices() helpers.ProductPrices
	GetReviewScore() string
	GetReleaseDateNice() string
	GetGroupID() string
}

type Player interface {
	GetName() string
	GetPath() string
	GetPathAbsolute() string
	GetAvatarAbsolute() string
	GetGamesCount() int
	GetAchievements() int
	GetPlaytime() int
	GetLevel() int
	GetBadges() int
	GetBadgesFoil() int
	GetRanks() map[helpers.RankMetric]int
	GetVACBans() int
	GetGameBans() int
	GetLastBan() time.Time
}

func getAppEmbed(commandID string, app App, code steamapi.ProductCC) *discordgo.MessageEmbed {

	var image string
	if app.GetPlayersPeakWeek() > 0 {
		image = charts.GetAppPlayersChart(commandID, app.GetID(), "1d", "180d", "Players (6 Months)")
	} else {
		image = charts.GetGroupChart(commandID, app.GetGroupID(), "Followers")
	}

	return &discordgo.MessageEmbed{
		Title:     app.GetName(),
		URL:       app.GetPathAbsolute(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage(), Width: 460, Height: 215},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Image:     &discordgo.MessageEmbedImage{URL: image},
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

func searchForPlayer(search string) (player elasticsearch.Player, err error) {

	// Check Elastic
	players, _, err := elasticsearch.SearchPlayers(1, 0, search, nil, nil)
	if err != nil {
		return player, err
	}

	if len(players) > 0 {
		return players[0], nil
	}

	// Check Steam
	tempPlayer, err := steam.GetPlayer(search)
	if err != nil {
		return player, err
	}

	player = elasticsearch.Player{
		ID:          tempPlayer.ID,
		PersonaName: tempPlayer.PersonaName,
		Avatar:      tempPlayer.Avatar,
		PlayTime:    tempPlayer.PlayTime,
		Games:       tempPlayer.Games,
		Level:       tempPlayer.Level,
		GameBans:    tempPlayer.GameBans,
		VACBans:     tempPlayer.VACBans,
		LastBan:     tempPlayer.LastBan.Unix(),
		// Friends:     tempPlayer.Friends,
	}

	// Queue
	err = consumers.ProducePlayer(consumers.PlayerMessage{ID: player.ID}, "chatbot-player")
	err = helpers.IgnoreErrors(err, consumers.ErrInQueue)
	if err != nil {
		log.Err("Producing player", zap.Error(err))
	}

	return player, nil
}
