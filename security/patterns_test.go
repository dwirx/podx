package security

import "testing"

func TestPatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"aws_access_key", "AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE", true},
		{"aws_secret", "aws_secret_access_key = 'wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY'", true},
		{"private_key", "-----BEGIN RSA PRIVATE KEY-----", true},
		{"password_assignment", `password = "secret123"`, true},
		{"password_single_quote", `password='secret123'`, true},
		{"api_key", `api_key = "sk-1234567890"`, true},
		{"generic_secret", `secret = "mysecret"`, true},
		{"connection_string", "mongodb://user:pass@localhost:27017", true},
		{"postgres_url", "postgres://admin:password@db.example.com/mydb", true},
		{"safe_code", "const username = 'admin'", false},
		{"safe_comment", "// TODO: add password validation", false},
		{"encrypted_value", "PASSWORD=ENC[age:YWdlLWVuY3J5cH...]", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContainsSecret(tt.content); got != tt.want {
				t.Errorf("ContainsSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindSecrets(t *testing.T) {
	content := `DB_HOST=localhost
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
password = "admin123"
DEBUG=true`

	matches := FindSecrets(content)
	if len(matches) != 2 {
		t.Errorf("FindSecrets() found %d matches, want 2", len(matches))
	}
}
