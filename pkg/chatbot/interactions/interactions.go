package interactions

//goland:noinspection GoUnusedConst
const (
	CommandOptionSubCommand = iota + 1
	CommandOptionSubCommandGroup
	CommandOptionString
	CommandOptionInteger
	CommandOptionBoolean
	CommandOptionUser
	CommandOptionChannel
	CommandOptionRole
)

type Interaction struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Options     []InteractionOption `json:"options"`
}

type InteractionOption struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Type        int                 `json:"type"`
	Required    bool                `json:"required"`
	Choices     []InteractionChoice `json:"choices,omitempty"`
}

type InteractionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
