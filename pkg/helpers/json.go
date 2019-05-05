package helpers

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/gamedb/gamedb/pkg/log"
)

var ErrUnMarshalNonPointer = errors.New("trying to unmarshal a non-pointer")

func IsJSON(str string) bool {
	var js json.RawMessage
	return Unmarshal([]byte(str), &js) == nil
}

func MarshalLog(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	log.Err(err)
	return bytes
}

// Wraps json.Unmarshal and adds logging
func Unmarshal(data []byte, v interface{}) (err error) {

	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return ErrUnMarshalNonPointer
	}

	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, v)

	switch err.(type) {
	case *json.SyntaxError, *json.InvalidUnmarshalError, *json.UnmarshalTypeError:
		if len(data) > 1000 {
			data = data[0:1000]
		}
		log.Info(err.Error() + ": " + string(data) + "...")
	default:
		log.Err(err)
	}

	return err
}
