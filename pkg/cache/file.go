package cache

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/djherbis/fscache"
	"github.com/gamedb/gamedb/pkg/log"
)

func GetSetCache(name string, ttl time.Duration, retrieve func() (interface{}, error), val interface{}) (err error) {

	if ttl == 0 {
		ttl = time.Hour * 24 * 365
	}

	c, err := fscache.New("./cache", 0755, ttl)
	if err != nil {
		return err
	}

	reader, writer, err := c.Get(name)
	if err != nil {
		return err
	}

	defer func() {
		err = reader.Close()
		log.Err(err)
	}()

	// Read from cache
	if writer == nil {
		dec := gob.NewDecoder(reader)
		return dec.Decode(val)
	}

	log.Info("Saving " + name + " to cache")

	// Write to cache
	defer func() {
		err = writer.Close()
		log.Err(err)
	}()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	i, err := retrieve()
	if err != nil {
		return err
	}

	err = encoder.Encode(i)
	if err != nil {
		return err
	}

	// Save to cache
	_, err = writer.Write(buf.Bytes())
	return err
}
