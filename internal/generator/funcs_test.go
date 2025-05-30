package generator

import (
	"os"
	"reflect"
	"testing"
)

func TestUnique_IntSlice(t *testing.T) {
	input := []int{1, 2, 2, 3, 1, 4}
	expected := []int{1, 2, 3, 4}
	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, ok := out.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", out)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestUnique_StringSlice(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	expected := []string{"a", "b", "c"}
	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, ok := out.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", out)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestUnique_EmptySlice(t *testing.T) {
	input := []int{}
	expected := []int{}
	out, err := unique(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, ok := out.([]int)
	if !ok {
		t.Fatalf("expected []int, got %T", out)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestUnique_NilInput(t *testing.T) {
	out, err := unique(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil, got %v", out)
	}
}

func TestUnique_NotSlice(t *testing.T) {
	_, err := unique(123)
	if err == nil {
		t.Fatal("expected error for non-slice input")
	}
}

func TestUnique_NonComparableElement(t *testing.T) {
	type foo struct {
		bar []int
	}
	input := []foo{{bar: []int{1}}, {bar: []int{1}}}
	_, err := unique(input)
	if err == nil {
		t.Fatal("expected error for non-comparable element type")
	}
}

func TestGetEnvOrDefault_NoEnv(t *testing.T) {
	const key = "NON_EXISTENT_ENV"
	os.Unsetenv(key)
	got := getEnvOrDefault(key, "defaultVal")
	if got != "defaultVal" {
		t.Errorf("expected defaultVal, got %q", got)
	}
}

func TestGetEnvOrDefault_WithEnv(t *testing.T) {
	const key = "EXISTENT_ENV"
	os.Setenv(key, "setVal")
	defer os.Unsetenv(key)
	got := getEnvOrDefault(key, "defaultVal")
	if got != "setVal" {
		t.Errorf("expected setVal, got %q", got)
	}
}
