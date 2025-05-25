package loader

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func LoadYaml(input []byte) (any, error) {
	var data any
	err := yaml.Unmarshal(input, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML input: %w", err)
	}

	return data, nil
}
