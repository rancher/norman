package types

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v2"
)

func JSONEncoder(writer io.Writer, v interface{}) error {
	return json.NewEncoder(writer).Encode(v)
}

func YamlEncoder(writer io.Writer, v interface{}) error {
	return yaml.NewEncoder(writer).Encode(v)
}
