package loader

import (
	"reflect"
	"testing"
)

func TestLoadYaml_Success(t *testing.T) {
	input := []byte("key: value\nnumber: 42")
	expected := map[string]interface{}{
		"key":    "value",
		"number": 42,
	}

	result, err := LoadYaml(input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The YAML library unmarshals into map[string]interface{}
	resMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", result)
	}

	if !reflect.DeepEqual(resMap, expected) {
		t.Errorf("expected %v, got %v", expected, resMap)
	}
}

func TestLoadYaml_InvalidYAML(t *testing.T) {
	input := []byte("key: value\n  -- invalid: hello")
	_, err := LoadYaml(input)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadYaml_EmptyInput(t *testing.T) {
	input := []byte("")
	result, err := LoadYaml(input)
	if err != nil {
		t.Fatalf("expected no error for empty input, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for empty input, got %v", result)
	}
}
