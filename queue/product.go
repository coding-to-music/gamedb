package queue

type RabbitMessageProduct struct {
	ID           int                           `json:"ID"`
	ChangeNumber int                           `json:"ChangeNumber"`
	MissingToken bool                          `json:"MissingToken"`
	SHAHash      string                        `json:"SHAHash"`
	KeyValues    RabbitMessageProductKeyValues `json:"KeyValues"`
	OnlyPublic   bool                          `json:"OnlyPublic"`
	UseHTTP      bool                          `json:"UseHttp"`
	HTTPURI      interface{}                   `json:"HttpUri"`
}

type RabbitMessageProductKeyValues struct {
	Name     string                          `json:"Name"`
	Value    interface{}                     `json:"Value"`
	Children []RabbitMessageProductKeyValues `json:"Children"`
}

type KeyValueMap map[string]KeyValueStruct

type KeyValueStruct struct {
	Name     string      `json:"Name"`
	Value    interface{} `json:"Value"`
	Children KeyValueMap `json:"Children"`
}

func (i RabbitMessageProductKeyValues) Convert() (o KeyValueStruct) {

	o.Name = i.Name
	o.Value = i.Value
	o.Children = KeyValueMap{}

	for _, v := range i.Children {
		o.Children[v.Name] = v.Convert()
	}

	return o
}
