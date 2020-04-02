package memcache

import (
	"encoding/json"
	"errors"
	"reflect"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/memcachier/mc"
)

var (
	client = mc.NewMC(config.Config.MemcacheDSN.Get(), config.Config.MemcacheUsername.Get(), config.Config.MemcachePassword.Get())

	ErrNotPointer = errors.New("value must be a pointer")
	ErrNotFound   = mc.ErrNotFound
)

func Delete(key string) (err error) {

	return client.Del(key)
}

func DeleteAll() error {

	return client.Flush(0)
}

func Get(key string) (val string, err error) {

	val, _, _, err = client.Get(key)
	return val, err
}

func Set(key string, val string, exp uint32) (err error) {

	_, err = client.Set(key, val, 0, exp, 0)
	return err
}

func GetInterface(key string, i interface{}) (err error) {

	val, _, _, err := client.Get(key)
	return json.Unmarshal([]byte(val), i)
}

func SetInterface(key string, val interface{}, exp uint32) (err error) {

	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = client.Set(key, string(b), 0, exp, 0)
	return err
}

func GetSetInterface(key string, exp uint32, value interface{}, callback func() (interface{}, error)) (err error) {

	if config.IsLocal() && reflect.TypeOf(value).Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	err = GetInterface(key, value)
	if err == mc.ErrNotFound {

		var s interface{}

		s, err = callback()
		if err != nil {
			return err
		}

		if config.IsLocal() && reflect.TypeOf(s) != reflect.TypeOf(value).Elem() {
			return errors.New(reflect.TypeOf(s).String() + " does not match " + reflect.TypeOf(value).Elem().String())
		}

		err = helpers.MarshalUnmarshal(s, value)
		if err != nil {
			return err
		}

		return SetInterface(key, s, exp)
	}

	return err
}
