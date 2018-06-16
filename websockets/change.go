package websockets

type Changes struct {
	Changes []Change `json:"id"`
}

func (c *Changes) AddChange(change Change) {
	c.Changes = append(c.Changes, change)
}

type Change struct {
	ID            int          `json:"id"`
	CreatedAtUnix int64        `json:"created_at"`
	CreatedAtNice string       `json:"created_at_nice"`
	Apps          []ChangeItem `json:"apps"`
	Packages      []ChangeItem `json:"packages"`
}

func (c *Change) AddApp(app ChangeItem) {
	c.Apps = append(c.Apps, app)
}

func (c *Change) AddPackage(pack ChangeItem) {
	c.Packages = append(c.Packages, pack)
}

type ChangeItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
