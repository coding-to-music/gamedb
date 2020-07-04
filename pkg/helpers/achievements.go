package helpers

import (
	"strconv"
	"strings"
)

func GetAchievementIcon(appID int, icon string) string {

	if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}

	if RegexSha1Only.MatchString(icon) {
		return AppIconBase + strconv.Itoa(appID) + "/" + icon + ".jpg"
	}

	return DefaultAppIcon
}

func GetAchievementCompleted(f float64) string {
	return FloatToString(f, 1)
}
