package function

import (
	b64 "encoding/base64"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

func read(reader io.ReadCloser) []byte {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil
	}
	return b
}

type (
	MergebotFile struct {
		Version string `yaml:"version"`
		Use     string `yaml:"use"`
	}
)

func UnmarshalMergebotFile(data []byte) (MergebotFile, error) {
	mf := MergebotFile{}
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return mf, err
	}
	return mf, nil
}

type (
	Config struct {
		ApplicationID         int
		ApplicationPrivateKey []byte
		WebhookSecret         string
	}
)

func BuildConfig() Config {
	appID, _ := strconv.Atoi(os.Getenv("APPLICATION_ID"))
	privateKey, err := decode(os.Getenv("APPLICATION_PRIVATE_KEY_B64"))
	if err != nil {
		panic(err)
	}
	secret := os.Getenv("WEBHOOK_SECRET")
	cnf := Config{
		ApplicationID:         appID,
		ApplicationPrivateKey: privateKey,
		WebhookSecret:         secret,
	}
	return cnf
}

func decode(in string) ([]byte, error) {
	uDec, err := b64.URLEncoding.DecodeString(in)
	return uDec, err
}
