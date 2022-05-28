package config

import (
	b64 "encoding/base64"
	"fmt"
	"os"
	"strconv"
)

type (
	Config struct {
		ApplicationID            int
		ApplicationPrivateKey    []byte
		WebhookSecret            string
		MarketplaceWebhookSecret string
		SkipPayloadValidation    bool
		DefaultRootCmd           string
	}
)

func BuildConfig() Config {
	appID, err := strconv.Atoi(getEnvOrDie("APPLICATION_ID"))
	if err != nil {
		panic(err)
	}
	privateKey, err := decode(getEnvOrDie("APPLICATION_PRIVATE_KEY_B64"))
	if err != nil {
		panic(err)
	}
	secret := os.Getenv("WEBHOOK_SECRET")
	marketplaceSecret := os.Getenv("MARKETPLACE_WEBHOOK_SECRET")
	skipPayloadValidation := os.Getenv("SKIP_PAYLOAD_VALIDATION") == "true"

	cnf := Config{
		ApplicationID:            appID,
		ApplicationPrivateKey:    privateKey,
		WebhookSecret:            secret,
		MarketplaceWebhookSecret: marketplaceSecret,
		SkipPayloadValidation:    skipPayloadValidation,
		DefaultRootCmd:           getEnv("ROOT_CMD", "pr-bot"),
	}
	return cnf
}

func decode(in string) ([]byte, error) {
	uDec, err := b64.URLEncoding.DecodeString(in)
	return uDec, err
}

func getEnvOrDie(name string) string {
	v := os.Getenv(name)
	if v == "" {
		panic(fmt.Errorf("%s is required", name))
	}
	return v
}

func getEnv(name string, fallback string) string {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}
	return v
}
