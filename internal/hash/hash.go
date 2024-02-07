package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

const Header = "HashSHA256"

func Encode(bytes []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(bytes)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
