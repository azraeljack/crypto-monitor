package config

import (
	"encoding/json"
)

type Config struct {
	Notifiers  []json.RawMessage `json:"notifiers"`
	Collectors []json.RawMessage `json:"collectors"`
	Strategies []json.RawMessage `json:"strategies"`
}
