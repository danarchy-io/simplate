package validator

import (
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func ValidateYamlWithSchema(input any, json []byte) error {

	schema, err := jsonschema.CompileString("schema.json", string(json))
	if err != nil {
		return fmt.Errorf("failed to compile JSONSchema: %w", err)
	}

	return schema.Validate(input)
}
