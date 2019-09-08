package pics

type Depots struct {
	Depots   []AppDepotItem
	Branches []AppDepotBranches
	Extra    map[string]string
}
type AppDepotItem struct {
	ID                         int               `json:"id"`
	Name                       string            `json:"name"`
	Configs                    map[string]string `json:"config"`
	Manifests                  map[string]string `json:"manifests"`
	EncryptedManifests         string            `json:"encryptedmanifests"`
	MaxSize                    uint64            `json:"maxsize"`
	App                        int               `json:"depotfromapp"`
	DLCApp                     int               `json:"dlcappid"`
	SystemDefined              bool              `json:"systemdefined"`
	Optional                   bool              `json:"optional"`
	SharedInstall              bool              `json:"sharedinstall"`
	SharedDepotType            bool              `json:"shareddepottype"`
	LVCache                    bool              `json:"lvcache"`
	AllowAddRemoveWhileRunning bool              `json:"allowaddremovewhilerunning"`
}
type AppDepotBranches struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	BuildID          int    `json:"buildid"`
	TimeUpdated      int64  `json:"timeupdated"`
	PasswordRequired bool   `json:"pwdrequired"`
	LCSRequired      bool   `json:"lcsrequired"`
	DefaultForSubs   string `json:"defaultforsubs"`
	UnlockForSubs    string `json:"unlockforsubs"`
}

type DepotConfig struct {
	OSList      string `json:"oslist"`
	OSArch      string `json:"osarch"`
	OptionalDLC string `json:"optionaldlc"`
	Language    string `json:"language"`
	LowViolence string `json:"lowviolence"`
}

type DepotManifest struct {
	Local  string `json:"local"`
	Public string `json:"public"`
}
