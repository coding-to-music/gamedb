package helpers

import (
	"strconv"
	"strings"
)

func GetAchievementIcon(appID int, icon string) string {

	if !strings.HasPrefix(icon, "/") && !strings.HasPrefix(icon, "http") {
		icon = AppIconBase + strconv.Itoa(appID) + "/" + icon
	}
	if !strings.HasSuffix(icon, ".jpg") {
		icon = icon + ".jpg"
	}

	return icon
}
