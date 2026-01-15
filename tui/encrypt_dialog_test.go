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
