package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

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

func decodeSha(secret string, in []byte) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(in)
	sha := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("sha256=%s", sha)
}
