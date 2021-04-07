package schemas

type PlayerField string

var (
	InfPlayersAchievements         PlayerField = "achievements"
	InfPlayersAchievements100      PlayerField = "achievements_count_100"
	InfPlayersAchievementsApps     PlayerField = "achievements_count_apps"
	InfPlayersBadges               PlayerField = "badges"
	InfPlayersBadgesFoil           PlayerField = "badges_foil"
	InfPlayersComments             PlayerField = "comments"
	InfPlayersFriends              PlayerField = "friends"
	InfPlayersGames                PlayerField = "games"
	InfPlayersLevel                PlayerField = "level"
	InfPlayersPlaytime             PlayerField = "playtime"
	InfPlayersAwardsGivenCount     PlayerField = "awards_given_count"
	InfPlayersAwardsGivenPoints    PlayerField = "awards_given_points"
	InfPlayersAwardsReceivedCount  PlayerField = "awards_received_count"
	InfPlayersAwardsReceivedPoints PlayerField = "awards_received_points"

	InfPlayersAchievementsRank         PlayerField = "achievements_rank"
	InfPlayersBadgesRank               PlayerField = "badges_rank"
	InfPlayersBadgesFoilRank           PlayerField = "badges_foil_rank"
	InfPlayersCommentsRank             PlayerField = "comments_rank"
	InfPlayersFriendsRank              PlayerField = "friends_rank"
	InfPlayersGamesRank                PlayerField = "games_rank"
	InfPlayersLevelRank                PlayerField = "level_rank"
	InfPlayersPlaytimeRank             PlayerField = "playtime_rank"
	InfPlayersAwardsGivenPointsRank    PlayerField = "awards_given_points_rank"
	InfPlayersAwardsReceivedPointsRank PlayerField = "awards_received_points_rank"
)
