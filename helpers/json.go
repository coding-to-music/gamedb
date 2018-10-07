package helpers

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/steam-authority/steam-authority/logger"
)

var ErrUnMarshalNonPointer = errors.New("trying to unmarshal a non-pointer")

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

			logger.Info(err.Error() + " - " + string(data) + "...")

		} else {
			logger.Error(err)
		}

		return err
	}

	return nil
}
