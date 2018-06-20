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

func (i RabbitMessageProductKeyValues) Convert() (o KeyValueStruct) {

	var valueString string
	if i.Value == nil {
		valueString = ""
	} else {
		valueString = i.Value.(string)
	}

	o.Name = i.Name
	o.Value = valueString
	o.Children = KeyValueMap{}

	for _, v := range i.Children {
		o.Children[v.Name] = v.Convert()
	}

	return o
}

type KeyValueMap map[string]KeyValueStruct

type KeyValueStruct struct {
	Name     string      `json:"Name"`
	Value    string      `json:"Value"`
	Children KeyValueMap `json:"Children"`
}

func (m KeyValueMap) GetValuesSlice() []string {
	var ret []string
	for _, v := range m {
		ret = append(ret, v.Value)
	}
	return ret
}

func (m KeyValueMap) GetValuesMap() map[string]string {
	var ret = map[string]string{}
	for _, v := range m {
		ret[v.Name] = v.Value
	}
	return ret
}
