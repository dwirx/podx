package tui

import (
	"testing"
)

// TestCreateFormForGPGKeygen verifies that the GPG keygen form is properly configured
func TestCreateFormForGPGKeygen(t *testing.T) {
	form := CreateFormForCommand("keygen-gpg")

	if form == nil {
		t.Fatal("CreateFormForCommand returned nil for keygen-gpg")
	}

	// Verify form title
	if form.Title != "Generate GPG Key" {
		t.Errorf("Expected title 'Generate GPG Key', got '%s'", form.Title)
	}

	// Verify form description
	if form.Description != "Generate a new GPG key pair" {
		t.Errorf("Expected description 'Generate a new GPG key pair', got '%s'", form.Description)
	}

	// Verify command ID
	if form.CommandID != "keygen-gpg" {
		t.Errorf("Expected CommandID 'keygen-gpg', got '%s'", form.CommandID)
	}

	// Verify number of fields
	if len(form.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(form.Fields))
	}

	// Verify first field (Name)
	nameField := form.Fields[0]
	if nameField.Label != "Name" {
		t.Errorf("Expected field[0].Label 'Name', got '%s'", nameField.Label)
	}
	if nameField.Placeholder != "Your Full Name" {
		t.Errorf("Expected field[0].Placeholder 'Your Full Name', got '%s'", nameField.Placeholder)
	}
	if !nameField.Required {
		t.Error("Expected field[0].Required to be true")
	}
	if nameField.Password {
		t.Error("Expected field[0].Password to be false")
	}

	// Verify second field (Email)
	emailField := form.Fields[1]
	if emailField.Label != "Email" {
		t.Errorf("Expected field[1].Label 'Email', got '%s'", emailField.Label)
	}
	if emailField.Placeholder != "you@example.com" {
		t.Errorf("Expected field[1].Placeholder 'you@example.com', got '%s'", emailField.Placeholder)
	}
	if !emailField.Required {
		t.Error("Expected field[1].Required to be true")
	}
	if emailField.Password {
		t.Error("Expected field[1].Password to be false")
	}

	// Verify inputs are initialized
	if len(form.inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(form.inputs))
	}
}

// TestBuildCommandArgsForGPGKeygen verifies that command args are built correctly
func TestBuildCommandArgsForGPGKeygen(t *testing.T) {
	values := map[string]string{
		"Name":  "John Doe",
		"Email": "john@example.com",
	}

	args := BuildCommandArgs("keygen-gpg", values)

	expectedArgs := []string{"keygen", "-t", "gpg", "-n", "John Doe", "-e", "john@example.com"}

	if len(args) != len(expectedArgs) {
		t.Fatalf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expected := range expectedArgs {
		if args[i] != expected {
			t.Errorf("Expected arg[%d] = '%s', got '%s'", i, expected, args[i])
		}
	}
}

// TestCreateFormForAllCommands verifies all commands with needsForm get a form
func TestCreateFormForAllCommands(t *testing.T) {
	commands := GetAllCommands()

	for _, item := range commands {
		if cmd, ok := item.(CommandItem); ok {
			if cmd.needsForm {
				t.Run(cmd.id, func(t *testing.T) {
					form := CreateFormForCommand(cmd.id)
					if form == nil {
						t.Errorf("Command %s has needsForm=true but CreateFormForCommand returned nil", cmd.id)
					}
				})
			}
		}
	}
}

// TestFormValidation tests form validation logic
func TestFormValidation(t *testing.T) {
	testCases := []struct {
		name      string
		commandID string
		values    map[string]string
		wantErr   bool
	}{
		{
			name:      "valid add-recipient",
			commandID: "add-recipient",
			values: map[string]string{
				"Name": "John Doe",
				"Key":  "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p",
			},
			wantErr: false,
		},
		{
			name:      "invalid add-recipient - bad key",
			commandID: "add-recipient",
			values: map[string]string{
				"Name": "John Doe",
				"Key":  "not-an-age-key",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			form := CreateFormForCommand(tc.commandID)
			if form == nil {
				t.Fatal("CreateFormForCommand returned nil")
			}

			// Set input values
			for i, field := range form.Fields {
				if val, ok := tc.values[field.Label]; ok {
					form.inputs[i].SetValue(val)
				}
			}

			err := form.validate()
			if tc.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}
