package influx

var (
	InfPlayersAchievements     = Field{Field: "achievements"}
	InfPlayersAchievementsRank = Field{Field: "achievements_rank"}
	InfPlayersBadges           = Field{Field: "badges"}
	InfPlayersBadgesRank       = Field{Field: "badges_rank"}
	InfPlayersBadgesFoil       = Field{Field: "badges_foil"}
	InfPlayersBadgesFoilRank   = Field{Field: "badges_foil_rank"}
	InfPlayersComments         = Field{Field: "comments"}
	InfPlayersCommentsRank     = Field{Field: "comments_rank"}
	InfPlayersFriends          = Field{Field: "friends"}
	InfPlayersFriendsRank      = Field{Field: "friends_rank"}
	InfPlayersGames            = Field{Field: "games"}
	InfPlayersGamesRank        = Field{Field: "games_rank"}
	InfPlayersLevel            = Field{Field: "level"}
	InfPlayersLevelRank        = Field{Field: "level_rank"}
	InfPlayersPlaytime         = Field{Field: "playtime"}
	InfPlayersPlaytimeRank     = Field{Field: "playtime_rank"}
)

type Field struct {
	Field    string
	Function string
}

func (c Field) String() string {

	if c.Function == "" {
		return c.Field
	}
	return c.Function + `("` + c.Field + `")`
}

func (c Field) Alias() string {
	return c.Function + "_" + c.Field
}

func (c *Field) Mean() *Field {
	c.Function = "mean"
	return c
}

func (c *Field) Median() *Field {
	c.Function = "median"
	return c
}

func (c *Field) Count() *Field {
	c.Function = "count"
	return c
}

func (c *Field) Min() *Field {
	c.Function = "min"
	return c
}

func (c *Field) Last() *Field {
	c.Function = "last"
	return c
}

func (c *Field) First() *Field {
	c.Function = "first"
	return c
}

func (c *Field) Sum() *Field {
	c.Function = "sum"
	return c
}

func (c *Field) Max() *Field {
	c.Function = "max"
	return c
}

func (c *Field) Spread() *Field {
	c.Function = "spread"
	return c
}

func (c *Field) StdDev() *Field {
	c.Function = "stddev"
	return c
}
