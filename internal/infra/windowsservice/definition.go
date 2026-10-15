package windowsservice

import "encoding/json"

// Definition describes a detached Windows service process.
// The definition is stored outside the service command line so executable
// arguments and paths do not need to be escaped into the Service Control
// Manager command string.
type Definition struct {
	ServiceName string   `json:"serviceName"`
	Binary      string   `json:"binary"`
	Arguments   []string `json:"arguments"`
	Directory   string   `json:"directory"`
	Log         string   `json:"log"`
}

func Encode(definition Definition) ([]byte, error) {
	return json.MarshalIndent(definition, "", "  ")
}
