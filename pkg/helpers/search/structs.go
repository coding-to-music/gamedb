package search

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type SearchResult struct {
	Keywords []string
	Name     string
	ID       uint64
	Icon     string
	Type     string
}

func (m SearchResult) GetKey() string {
	return m.Type + "-" + strconv.FormatUint(m.ID, 10)
}

func (m SearchResult) GetIcon() string {
	if m.Type == SearchTypeApp {
		return helpers.GetAppIcon(int(m.ID), m.Icon)
	} else if m.Type == SearchTypePlayer {
		return helpers.GetPlayerAvatar(m.Icon)
	}
	return ""
}
