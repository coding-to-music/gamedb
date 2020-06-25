package helpers

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/gamedb/gamedb/pkg/log"
)

var ErrUnMarshalNonPointer = errors.New("trying to unmarshal a non-pointer")

func IsJSON(str string) bool {
	var js json.RawMessage
	return Unmarshal([]byte(str), &js) == nil
}

// Wraps json.Unmarshal and adds logging
func Unmarshal(data []byte, v interface{}) (err error) {

	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, v)

	switch err.(type) {
	case *json.SyntaxError, *json.InvalidUnmarshalError, *json.UnmarshalTypeError:
		if len(data) > 1000 {
			data = data[0:1000]
		}
		log.Err(err.Error(), string(data))
	default:
		log.Err(err)
	}

	return err
}

// func UnmarshalNumber(b []byte, v interface{}) (err error) {
//
// 	d := json.NewDecoder(bytes.NewReader(b))
// 	d.UseNumber()
//
// 	return d.Decode(&v)
// }

func UnmarshalStrict(data []byte, v interface{}) error {

	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func FormatJSON(unformatted string) (formatted string, err error) {

	var x interface{}
	err = json.Unmarshal([]byte(unformatted), &x)
	if err != nil {
		return formatted, err
	}

	b, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		return formatted, err
	}

	return string(b), err
}

func MarshalUnmarshal(in interface{}, out interface{}) (err error) {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func FlattenMap(m map[string]interface{}) map[string]interface{} {

	o := make(map[string]interface{})
	for k, v := range m {

		switch child := v.(type) {
		case map[string]interface{}:
			nm := FlattenMap(child)
			for nk, nv := range nm {
				o[k+" / "+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}
