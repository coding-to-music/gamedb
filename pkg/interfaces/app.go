package interfaces

import (
	"github.com/gamedb/gamedb/pkg/helpers"
)

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
