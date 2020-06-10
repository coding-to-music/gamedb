package helpers

import (
	"strconv"
	"strings"
)

func GetAchievementIcon(appID int, icon string) string {

	if !strings.HasSuffix(icon, ".jpg") {
		icon = icon + ".jpg"
	}

	if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}

	// Return app
	return AppIconBase + strconv.Itoa(appID) + "/" + icon
}
