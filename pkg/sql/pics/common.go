package pics

import (
	"github.com/Jleagle/unmarshal-go/ctypes"
)

type Associations map[string]struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LibraryAssets struct {
	LibraryCapsule string `json:"library_capsule"`
	LibraryHero    string `json:"library_hero"`
	LibraryLogo    string `json:"library_logo"`
	LogoPosition   struct {
		HeightPct      string `json:"height_pct"`
		PinnedPosition string `json:"pinned_position"`
		WidthPct       string `json:"width_pct"`
	} `json:"logo_position"`
}

type EULA map[string]struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type SupportedLanguages map[string]struct {
	Supported ctypes.CBool `json:"supported"`
}
