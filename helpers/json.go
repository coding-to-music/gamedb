package helpers

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/gamedb/website/log"
)

var ErrUnMarshalNonPointer = errors.New("trying to unmarshal a non-pointer")

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func MarshalLog(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	log.Log(err)
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
	if err != nil {

		if strings.Contains(err.Error(), "cannot unmarshal") {

			if len(data) > 1000 {
				data = data[0:1000]
			}

			log.Info(err.Error()+" - "+string(data)+"...")

		} else {
			log.Log(err)
		}
	}

	return err
}
