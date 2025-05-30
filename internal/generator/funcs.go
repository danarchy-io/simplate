package generator

import (
	"fmt"
	"os"
	"reflect"
)

// unique returns a new slice containing only the distinct elements from the provided slice.
// It preserves the order of first occurrence. Behavior:
//   - If input is nil, returns (nil, nil).
//   - If input is not a slice, returns an error.
//   - If the slice is empty, returns a zero‐length slice of the same element type.
//   - If element type is not comparable, returns an error.
//
// Parameters:
//   - input: any value expected to be a slice.
//
// Returns:
//   - any: a slice of the same type containing unique elements.
//   - error: non‐nil if input is invalid or element type is not comparable.
func unique(input any) (any, error) {
	if input == nil {
		return nil, nil
	}

	sliceVal := reflect.ValueOf(input)
	if sliceVal.Kind() != reflect.Slice {
		return nil, fmt.Errorf("unique: expected a slice, got %T", input)
	}

	if sliceVal.Len() == 0 {
		elemType := sliceVal.Type().Elem()
		return reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0).Interface(), nil
	}

	elemType := sliceVal.Type().Elem()
	if !elemType.Comparable() {
		return nil, fmt.Errorf("unique: elements of type %s are not comparable", elemType.String())
	}

	seen := make(map[any]bool)
	resultSlice := reflect.MakeSlice(sliceVal.Type(), 0, sliceVal.Len())

	for i := range sliceVal.Len() {
		element := sliceVal.Index(i)
		elementInterface := element.Interface()

		if !seen[elementInterface] {
			seen[elementInterface] = true
			resultSlice = reflect.Append(resultSlice, element)
		}
	}
	return resultSlice.Interface(), nil
}

// getEnvOrDefault returns the value of the environment variable named by key.
// If the variable is unset or its value is empty, defaultValue is returned.
//
// Parameters:
//   - key: the name of the environment variable to retrieve.
//   - defaultValue: the value to return if the environment variable is not set.
//
// Returns:
//   - string: the environment variable’s value, or defaultValue if unset or empty.
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
