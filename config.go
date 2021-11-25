package function

import (
	b64 "encoding/base64"
	"os"
	"strconv"
)

type (
	Config struct {
		ApplicationID         int
		ApplicationPrivateKey []byte
	}
)

func BuildConfig() Config {
	appID, _ := strconv.Atoi(os.Getenv("APPLICATION_ID"))
	privateKey, err := decode(os.Getenv("APPLICATION_PRIVATE_KEY_B64"))
	if err != nil {
		panic(err)
	}
	cnf := Config{
		ApplicationID:         appID,
		ApplicationPrivateKey: privateKey,
	}
	return cnf
}

func decode(in string) ([]byte, error) {
	uDec, err := b64.URLEncoding.DecodeString(in)
	return uDec, err
}
