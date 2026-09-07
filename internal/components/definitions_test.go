package components

import (
	"testing"
)

func TestValidateType(t *testing.T) {
	tests := []struct {
		value        any
		expectedType string
		fieldName    string
		expectError  bool
	}{
		{"hello", "string", "name", false},
		{42, "number", "count", false},
		{true, "boolean", "enabled", false},
		{[]any{"a", "b"}, "list", "items", false},
		{map[string]any{"key": "val"}, "object", "data", false},
		{"2024-01-01", "date", "created", false},
		{"<b>html</b>", "html", "content", false},
		{123, "string", "name", false},
		{[]any{1, 2}, "string", "name", true},
		{"text", "boolean", "flag", true},
		{"not a list", "list", "items", true},
		{"not an object", "object", "data", true},
	}

	for _, tt := range tests {
		err := validateType(tt.value, tt.expectedType, "testComponent", tt.fieldName)
		if tt.expectError && err == nil {
			t.Errorf("expected error for %s(%T) as %s, got nil",
				tt.fieldName, tt.value, tt.expectedType)
		}
		if !tt.expectError && err != nil {
			t.Errorf("unexpected error for %s(%T) as %s: %v",
				tt.fieldName, tt.value, tt.expectedType, err)
		}
	}
}

func TestDefinitionValidateOptions(t *testing.T) {
	def := Definition{
		Name: "test",
		Options: map[string]OptionDefinition{
			"title":    {Type: "string", Required: true},
			"count":    {Type: "number", Required: false},
			"optional": {Type: "string", Required: false},
		},
	}

	errs := def.ValidateOptions(Instance{
		Name: "test",
		Options: map[string]any{
			"title": "Hello",
			"count": 42,
		},
	})

	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}

	errs = def.ValidateOptions(Instance{
		Name: "test",
		Options: map[string]any{
			"count": "not a number",
		},
	})

	if len(errs) < 1 {
		t.Errorf("expected at least 1 error, got %d", len(errs))
	}
}

func TestNormalizeComponentName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"navbar", "navbar"},
		{"navbar.html", "navbar"},
		{"hero", "hero"},
	}

	for _, tt := range tests {
		result := NormalizeComponentName(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeComponentName(%s): expected %s, got %s",
				tt.input, tt.expected, result)
		}
	}
}

func TestIsValidPosition(t *testing.T) {
	tests := []struct {
		position  string
		isValid   bool
	}{
		{"top", true},
		{"bottom", true},
		{"left", true},
		{"right", true},
		{"center", true},
		{"TOP", true},
		{"Left", true},
		{"invalid", false},
		{"middle", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidPosition(tt.position)
		if result != tt.isValid {
			t.Errorf("IsValidPosition(%s): expected %v, got %v",
				tt.position, tt.isValid, result)
		}
	}
}

func TestResolvePosition(t *testing.T) {
	tests := []struct {
		name               string
		defPosition        string
		instancePosition   string
		expectedResolution string
	}{
		{"instance override", "left", "right", "right"},
		{"use definition default", "top", "", "top"},
		{"default to center", "", "", "center"},
		{"instance overrides definition", "left", "bottom", "bottom"},
	}

	for _, tt := range tests {
		def := Definition{
			Name:     "test",
			Position: tt.defPosition,
		}
		instance := Instance{
			Name:     "test",
			Position: tt.instancePosition,
		}
		result := def.ResolvePosition(instance)
		if result != tt.expectedResolution {
			t.Errorf("%s: expected %s, got %s", tt.name, tt.expectedResolution, result)
		}
	}
}

func TestValidateUnknownOptions(t *testing.T) {
	def := Definition{
		Name: "test",
		Options: map[string]OptionDefinition{
			"title": {Type: "string", Required: true},
			"count": {Type: "number", Required: false},
		},
	}

	errs := def.ValidateOptions(Instance{
		Name: "test",
		Options: map[string]any{
			"title":   "Hello",
			"unknown": "value",
		},
	})

	// Should have error for unknown option
	if len(errs) < 1 {
		t.Errorf("expected error for unknown option, got none")
	}

	foundUnknownError := false
	for _, err := range errs {
		if contains(err.Error(), "unknown option") {
			foundUnknownError = true
		}
	}
	if !foundUnknownError {
		t.Errorf("expected unknown option error, got: %v", errs)
	}
}

func TestValidatePosition(t *testing.T) {
	def := Definition{
		Name: "test",
		Options: map[string]OptionDefinition{},
	}

	errs := def.ValidateOptions(Instance{
		Name:     "test",
		Position: "invalid",
		Options:  map[string]any{},
	})

	// Should have error for invalid position
	if len(errs) < 1 {
		t.Errorf("expected error for invalid position, got none")
	}

	foundPositionError := false
	for _, err := range errs {
		if contains(err.Error(), "invalid position") {
			foundPositionError = true
		}
	}
	if !foundPositionError {
		t.Errorf("expected invalid position error, got: %v", errs)
	}
}

func contains(s, substr string) bool {
	// Simple substring check
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
