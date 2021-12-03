package function

import (
	"crypto/hmac"
	"crypto/sha256"
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
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
	PrBotFile struct {
		Version string `yaml:"version"`
		Use     string `yaml:"use"`
	}
)

func UnmarshalPrBotFile(data []byte) (PrBotFile, error) {
	mf := PrBotFile{}
	if err := yaml.Unmarshal(data, &mf); err != nil {
		return mf, err
	}
	return mf, nil
}

type (
	Config struct {
		ApplicationID            int
		ApplicationPrivateKey    []byte
		WebhookSecret            string
		MarketplaceWebhookSecret string
	}
)

func BuildConfig() Config {
	appID, _ := strconv.Atoi(os.Getenv("APPLICATION_ID"))
	privateKey, err := decode(os.Getenv("APPLICATION_PRIVATE_KEY_B64"))
	if err != nil {
		panic(err)
	}
	secret := os.Getenv("WEBHOOK_SECRET")
	marketplaceSecret := os.Getenv("MARKETPLACE_WEBHOOK_SECRET")
	cnf := Config{
		ApplicationID:            appID,
		ApplicationPrivateKey:    privateKey,
		WebhookSecret:            secret,
		MarketplaceWebhookSecret: marketplaceSecret,
	}
	return cnf
}

func decode(in string) ([]byte, error) {
	uDec, err := b64.URLEncoding.DecodeString(in)
	return uDec, err
}

func decodeSha(secret string, in []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(in)
	sha := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha256=%s", sha)
}
