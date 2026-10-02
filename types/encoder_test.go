package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestJSONToYAML(t *testing.T) {
	tests := map[string]struct {
		json string
		want string
	}{
		"empty object": {
			json: `{}`,
			want: "{}\n",
		},
		"empty array": {
			json: `[]`,
			want: "[]\n",
		},
		"null": {
			json: `null`,
			want: "null\n",
		},
		"scalar string": {
			json: `"hello"`,
			want: "hello\n",
		},
		"scalar number": {
			json: `42`,
			want: "42\n",
		},
		"keys are sorted": {
			json: `{"b":1,"a":2}`,
			want: "a: 2\nb: 1\n",
		},
		"nested maps and lists": {
			json: `{"nested":{"list":[1,"two",{"three":3}]}}`,
			want: "nested:\n  list:\n  - 1\n  - two\n  - three: 3\n",
		},
		"strings that look like other types are quoted": {
			json: `{"s":"true","n":"123","e":"","nil":null,"bool":false}`,
			want: "bool: false\ne: \"\"\n\"n\": \"123\"\nnil: null\ns: \"true\"\n",
		},
		"YAML 1.1 boolean-like strings are quoted": {
			json: `{"k":"yes","o":"on","y":"y"}`,
			want: "k: \"yes\"\no: \"on\"\n\"y\": \"y\"\n",
		},
		"numbers keep their precision": {
			json: `{"big":12345678901234567890,"f":1.5,"exp":1e3,"i64":9007199254740993}`,
			want: "big: 12345678901234567890\nexp: 1000\nf: 1.5\ni64: 9007199254740993\n",
		},
		"multiline strings use literal block style": {
			json: `{"multi":"line1\nline2"}`,
			want: "multi: |-\n  line1\n  line2\n",
		},
		"unicode is preserved": {
			json: `{"unicode":"héllo ☃"}`,
			want: "unicode: héllo ☃\n",
		},
		"commenter keys are emitted unquoted": {
			json: `{"zzz#(a)(b)name":"x"}`,
			want: "zzz#(a)(b)name: x\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := jsonToYAML([]byte(tt.json))
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestJSONToYAMLInvalidInput(t *testing.T) {
	_, err := jsonToYAML([]byte(`{bad`))
	assert.ErrorContains(t, err, "yaml: line 1: did not find expected")
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestYAMLEncoder(t *testing.T) {
	type inner struct {
		Port int `json:"port"`
	}
	type sample struct {
		Name     string            `json:"name"`
		Empty    string            `json:"empty,omitempty"`
		Hidden   string            `json:"-"`
		Labels   map[string]string `json:"labels"`
		Items    []inner           `json:"items"`
		Nil      *inner            `json:"nil"`
		Created  time.Time         `json:"created"`
		RawValue json.RawMessage   `json:"raw"`
	}

	tests := map[string]struct {
		value any
		want  string
	}{
		"nil": {
			value: nil,
			want:  "null\n",
		},
		"map": {
			value: map[string]any{"b": "two", "a": 1},
			want:  "a: 1\nb: two\n",
		},
		"slice of maps": {
			value: []map[string]int{{"a": 1}, {"b": 2}},
			want:  "- a: 1\n- b: 2\n",
		},
		"struct honours json tags": {
			value: sample{
				Name:     "test",
				Hidden:   "secret",
				Labels:   map[string]string{"app": "web"},
				Items:    []inner{{Port: 80}, {Port: 443}},
				Created:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
				RawValue: json.RawMessage(`{"x":[true]}`),
			},
			want: "created: \"2026-01-02T03:04:05Z\"\n" +
				"items:\n" +
				"- port: 80\n" +
				"- port: 443\n" +
				"labels:\n" +
				"  app: web\n" +
				"name: test\n" +
				"nil: null\n" +
				"raw:\n" +
				"  x:\n" +
				"  - true\n",
		},
		"html characters are not escaped": {
			value: map[string]string{"html": "<a href=\"x\">&</a>"},
			want:  "html: <a href=\"x\">&</a>\n",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, YAMLEncoder(&buf, tt.value))
			assert.Equal(t, tt.want, buf.String())
		})
	}
}

func TestYAMLEncoderRoundTrip(t *testing.T) {
	in := map[string]any{
		"string": "value",
		"number": 1.5,
		"bool":   true,
		"null":   nil,
		"list":   []any{"a", 2.0, false},
		"nested": map[string]any{"multi": "line1\nline2", "quoted": "123"},
	}

	var buf bytes.Buffer
	require.NoError(t, YAMLEncoder(&buf, in))

	var out map[string]any
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &out))
	// Normalise numeric types by passing both through JSON.
	wantJSON, err := json.Marshal(in)
	require.NoError(t, err)
	gotJSON, err := json.Marshal(out)
	require.NoError(t, err)
	assert.JSONEq(t, string(wantJSON), string(gotJSON))
}

func TestYAMLEncoderMarshalError(t *testing.T) {
	var buf bytes.Buffer
	err := YAMLEncoder(&buf, map[string]any{"ch": make(chan int)})
	var unsupported *json.UnsupportedTypeError
	assert.ErrorAs(t, err, &unsupported)
	assert.Empty(t, buf.String())
}

func TestYAMLEncoderWriteError(t *testing.T) {
	err := YAMLEncoder(errWriter{}, map[string]string{"a": "b"})
	assert.EqualError(t, err, "write failed")
}
