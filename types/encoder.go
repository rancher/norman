package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	"go.yaml.in/yaml/v3"
)

var (
	commenter = regexp.MustCompile(`(?m)^( *)zzz#\\((.*)\\)\\((.*)\\)([a-z]+.*):(.*)`)
)

func JSONEncoder(writer io.Writer, v any) error {
	return json.NewEncoder(writer).Encode(v)
}

// YAMLEncoder encodes a Go value into YAML format and writes it to the provided writer.
func YAMLEncoder(writer io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	buf, err := jsonToYAML(data)
	if err != nil {
		return err
	}
	buf = commenter.ReplaceAll(buf, []byte("${1}# ${4}:${5}"))
	_, err = writer.Write(buf)
	return err
}

func jsonToYAML(j []byte) ([]byte, error) {
	var jsonObj any
	err := yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling the YAML: %w", err)
	}

	// Match the output format of sigs.k8s.io/yaml: 2 space indents with
	// sequences not indented under their parent key.
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	enc.CompactSeqIndent()

	if err := enc.Encode(jsonObj); err != nil {
		return nil, fmt.Errorf("encoding the YAML: %w", err)
	}

	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("closing the YAML encoder: %w", err)
	}

	return buf.Bytes(), nil
}
