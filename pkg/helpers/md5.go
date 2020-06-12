package helpers

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(b []byte) string {

	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}
