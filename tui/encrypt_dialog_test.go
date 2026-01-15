package tui

import "testing"

// TestEncryptMethodConstants verifies that the EncryptMethod constants have the correct values
func TestEncryptMethodConstants(t *testing.T) {
	tests := []struct {
		name     string
		method   EncryptMethod
		expected int
	}{
		{
			name:     "MethodPassword should be 0",
			method:   MethodPassword,
			expected: 0,
		},
		{
			name:     "MethodAgeKey should be 1",
			method:   MethodAgeKey,
			expected: 1,
		},
		{
			name:     "MethodGPG should be 2",
			method:   MethodGPG,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.method) != tt.expected {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expected, int(tt.method))
			}
		})
	}
}

// TestEncryptDialogMethodSelection verifies the method selection in the encryption dialog
func TestEncryptDialogMethodSelection(t *testing.T) {
	// Create a new dialog model
	dialog := NewEncryptDialogModel()

	// Test 1: Verify default method is MethodPassword
	if dialog.method != MethodPassword {
		t.Errorf("expected default method to be MethodPassword (0), got %d", dialog.method)
	}

	// Test 2: Verify methods array has 3 elements
	expectedMethodCount := 3
	if len(dialog.methods) != expectedMethodCount {
		t.Errorf("expected methods array to have %d elements, got %d", expectedMethodCount, len(dialog.methods))
	}

	// Test 3: Verify method names are correct
	expectedMethods := []string{"Password", "Age Key", "GPG Key"}
	for i, expectedMethod := range expectedMethods {
		if i >= len(dialog.methods) {
			t.Errorf("methods array too short: expected at least %d elements", i+1)
			break
		}
		if dialog.methods[i] != expectedMethod {
			t.Errorf("expected methods[%d] to be %q, got %q", i, expectedMethod, dialog.methods[i])
		}
	}

	// Test 4: Verify methodStep is initialized (should start with method selection)
	if !dialog.methodStep {
		t.Error("expected methodStep to be true (start with method selection)")
	}
}
