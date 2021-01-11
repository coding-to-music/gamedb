package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func MD5(b []byte) string {

	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}

func MD5Interface(i interface{}) string {

	b, _ := json.Marshal(i)
	return MD5(b)
}
