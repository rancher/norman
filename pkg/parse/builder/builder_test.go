package builder

import (
	"testing"

	"github.com/rancher/norman/v2/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestEmptyStringWithDefault(t *testing.T) {
	schema := &types.Schema{
		ResourceFields: map[string]types.Field{
			"foo": {
				Default: "foo",
				Type:    "string",
				Create:  true,
			},
		},
	}
	schemas := types.EmptySchemas()
	schemas.AddSchema(*schema)

	builder := NewBuilder(&types.APIRequest{})

	// Test if no field we set to "foo"
	result, err := builder.Construct(schema, types.ToAPI(nil), Create)
	if err != nil {
		t.Fatal(err)
	}
	value, ok := result.Map()["foo"]
	assert.True(t, ok)
	assert.Equal(t, "foo", value)

	// Test if field is "" we set to "foo"
	result, err = builder.Construct(schema, types.ToAPI(map[string]interface{}{
		"foo": "",
	}), Create)
	if err != nil {
		t.Fatal(err)
	}
	value, ok = result.Map()["foo"]
	assert.True(t, ok)
	assert.Equal(t, "foo", value)
}
