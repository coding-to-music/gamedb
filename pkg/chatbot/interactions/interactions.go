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
	ID            string              `json:"id"`             // no need to set
	ApplicationID string              `json:"application_id"` // no need to set
	Name          string              `json:"name"`           // ^[\w-]{3,32}$
	Description   string              `json:"description"`    //
	Options       []InteractionOption `json:"options"`        //
}

type InteractionOption struct {
	Name        string              `json:"name"`              // ^[\w-]{1,32}$
	Description string              `json:"description"`       //
	Type        int                 `json:"type"`              //
	Required    bool                `json:"required"`          //
	Choices     []InteractionChoice `json:"choices,omitempty"` //
}

type InteractionChoice struct {
	Name  string `json:"name"`  // 1-100 character choice name
	Value string `json:"value"` // value of the choice
}
