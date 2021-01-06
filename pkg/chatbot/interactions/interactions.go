package interactions

//goland:noinspection GoUnusedConst
const (
	InteractionOptionTypeSubCommand = iota + 1
	InteractionOptionTypeSubCommandGroup
	InteractionOptionTypeString
	InteractionOptionTypeInteger
	InteractionOptionTypeBoolean
	InteractionOptionTypeUser
	InteractionOptionTypeChannel
	InteractionOptionTypeRole
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
