package template

import (
	"os"
	"reflect"
	"testing"
)

func TestUnique_IntSlice(t *testing.T) {
	input := []any{1, 2, 2, 3, 1, 4}
	expected := []any{1, 2, 3, 4}

	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("expected %v, got %v", expected, out)
	}
}

func TestUnique_StringSlice(t *testing.T) {
	input := []any{"a", "b", "a", "c", "b"}
	expected := []any{"a", "b", "c"}

	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("expected %v, got %v", expected, out)
	}
}

func TestUnique_EmptySlice(t *testing.T) {
	input := []any{}
	expected := []any{}

	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("expected %v, got %v", expected, out)
	}
}

func TestUnique_NilInput(t *testing.T) {
	var input []any
	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil, got %v", out)
	}
}

func TestUnique_NonComparableElement(t *testing.T) {
	type foo struct{ bar []int }
	// two distinct foo values, but slice field makes foo non-comparable
	f1 := foo{bar: []int{1}}
	f2 := foo{bar: []int{1}}
	input := []any{f1, f2}

	_, err := unique(input)
	if err == nil {
		t.Fatal("expected error for non-comparable element type")
	}
}

func TestGetEnvOrDefault_NoEnv(t *testing.T) {
	const key = "NON_EXISTENT_ENV"
	os.Unsetenv(key)
	got := envOrDefault(key, "defaultVal")
	if got != "defaultVal" {
		t.Errorf("expected defaultVal, got %q", got)
	}
}

func TestGetEnvOrDefault_WithEnv(t *testing.T) {
	const key = "EXISTENT_ENV"
	os.Setenv(key, "setVal")
	defer os.Unsetenv(key)
	got := envOrDefault(key, "defaultVal")
	if got != "setVal" {
		t.Errorf("expected setVal, got %q", got)
	}
}
