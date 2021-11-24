package config

import (
	"io/ioutil"
	"os"
	"strconv"
)

type (
	Config struct {
		ApplicationID         int
		ApplicationPrivateKey []byte
	}
)

func Build() Config {
	appID, _ := strconv.Atoi(os.Getenv("APPLICATION_ID"))
	privateKey, err := ioutil.ReadFile(os.Getenv("APPLICATION_PRIVATE_KEY_PATH"))
	if err != nil {
		panic(err)
	}
	cnf := Config{
		ApplicationID:         appID,
		ApplicationPrivateKey: privateKey,
	}
	return cnf
}
