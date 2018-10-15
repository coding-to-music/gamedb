package helpers

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/steam-authority/steam-authority/logging"
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

		if err2, ok := err.(*json.UnmarshalTypeError); ok {
			if SliceHasString([]string{"[]db.ProfileFriend", "[]db.ProfileBadge"}, err2.Type.String()) {
				logging.ErrorG(err2)
				return nil
			}
		}

		if strings.Contains(err.Error(), "cannot unmarshal") {

			if len(data) > 1000 {
				data = data[0:1000]
			}

			logging.Info(err.Error() + " - " + string(data) + "...")

		} else {
			logging.Error(err)
		}
	}

	return err
}
