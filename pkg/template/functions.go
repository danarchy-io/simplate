package template

import (
	"fmt"
	"os"
	"reflect"
)

// unique returns a new []any containing only the distinct elements from the provided slice.
// It preserves the order of first occurrence.
// Behavior:
//   - If input is nil, returns (nil, nil).
//   - If any element’s dynamic type is not comparable, returns an error.
//
// Parameters:
//   - input: []any
//
// Returns:
//   - []any: a slice containing unique elements in original order.
//   - error: non‐nil if any element is not comparable.
func unique(input []any) ([]any, error) {
	if input == nil {
		return nil, nil
	}
	seen := make(map[any]struct{}, len(input))
	result := make([]any, 0, len(input))

	for _, elem := range input {
		if elem != nil {
			t := reflect.TypeOf(elem)
			if !t.Comparable() {
				return nil, fmt.Errorf("unique: elements of type %s are not comparable", t.String())
			}
		}
		if _, exists := seen[elem]; !exists {
			seen[elem] = struct{}{}
			result = append(result, elem)
		}
	}
	return result, nil
}

// envOrDefault returns the value of the environment variable named by key.
// If the variable is unset or its value is empty, defaultValue is returned.
//
// Parameters:
//   - key: the name of the environment variable to retrieve.
//   - defaultValue: the value to return if the environment variable is not set.
//
// Returns:
//   - string: the environment variable’s value, or defaultValue if unset or empty.
func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
