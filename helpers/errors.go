package helpers

import "github.com/gamedb/website/logging"

// Ignore returns nil if an error is of one of the provided types, returns the provided error otherwise.
func IgnoreErrors(err error, errs ...error) error {

	if len(errs) == 0 {
		panic("Using IgnoreErrors wrong")
	}

	for _, v := range errs {
		if err == v {
			logging.ErrorL(err)
			return nil
		}
	}

	return err
}
