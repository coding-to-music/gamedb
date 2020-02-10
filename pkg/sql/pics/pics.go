package pics

import (
	"sort"
)

type PICSKeyValues map[string]string

func (kv PICSKeyValues) GetValue(key string) string {

	if val, ok := kv[key]; ok {
		return val
	}
	return ""
}

func (kv PICSKeyValues) Map() (m map[string]string) {
	m = map[string]string{}
	for k, v := range kv {
		m[k] = v
	}
	return m
}

func (kv PICSKeyValues) Formatted(productID int, keys map[string]PicsKey) (ret []KeyValue) {

	for k, v := range kv {
		ret = append(ret, KeyValue{
			Key:            k,
			Value:          v,
			ValueFormatted: FormatVal(k, v, productID, keys),
			Type:           getType(k, keys),
			Description:    getDescription(k, keys),
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret
}

type KeyValue struct {
	Key            string
	Value          string
	ValueFormatted interface{}
	Type           int
	Description    string
}

func (kv KeyValue) TDClass() string {

	switch kv.Type {
	case picsTypeImage:
		return "img"
	default:
		return ""
	}
}
