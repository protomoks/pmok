package config

import (
	"encoding/json"
)

type Function struct {
	HttpPathname   string   `json:"path" yaml:"path"`
	Entrypoint     string   `json:"entrypoint" yaml:"entrypoint"`
	AllowedMethods []string `json:"methods" yaml:"methods"`
}

type FunctionConfig map[string]Function

func (f FunctionConfig) ToJSON() ([]byte, error) {
	return json.Marshal(&f)
}
