package refreshtoken

import (
  "encoding/base64"
  "math/rand"
)

const letters  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"


func GetRefreshToken() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}

	return base64.StdEncoding.EncodeToString(b)
}
