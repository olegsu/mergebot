package functions

import (
	yaml "gopkg.in/yaml.v3"
)

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
