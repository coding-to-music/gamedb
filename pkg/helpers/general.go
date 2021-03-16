package helpers

import (
	"io"

	"github.com/gamedb/gamedb/pkg/log"
)

type Tuple struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type TupleInt struct {
	Key   int `json:"k"`
	Value int `json:"v"`
}

type TupleStringInt struct {
	Key   string `json:"k"`
	Value int64  `json:"v"`
}

// IgnoreErrors returns nil if an error is of one of the provided types, returns the provided error otherwise.
func IgnoreErrors(err error, errs ...error) error {

	for _, v := range errs {
		if err == v {
			return nil
		}
	}

	return err
}

func Close(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.ErrS(err)
	}
}

func StringsToInterfaces(s []string) (o []interface{}) {
	for _, v := range s {
		o = append(o, v)
	}
	return o
}

func IntsToInt32s(s []int) (o []int32) {
	for _, v := range s {
		o = append(o, int32(v))
	}
	return o
}

func Int32sToInts(s []int32) (o []int) {
	for _, v := range s {
		o = append(o, int(v))
	}
	return o
}
