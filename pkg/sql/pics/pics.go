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

func (kv PICSKeyValues) Formatted(productID int, keys map[string]PicsKey) (ret []KeyValue, err error) {

	for k, v := range kv {
		ret = append(ret, KeyValue{
			Key:            k,
			Value:          v,
			ValueFormatted: FormatVal(k, v, productID, keys),
			Type:           GetType(k, keys),
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		return i > j
	})

	return ret, err
}
