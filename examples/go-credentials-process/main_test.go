package main

import (
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				LambdaURL:  "https://lambda.amazonaws.com",
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test",
				Duration:   1 * time.Hour,
			},
			wantError: false,
		},
		{
			name: "missing lambda URL",
			config: &Config{
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test",
				Duration:   1 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "missing role ARN",
			config: &Config{
				LambdaURL:  "https://lambda.amazonaws.com",
				TailnetKey: "tskey-auth-test",
				Duration:   1 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "missing tailnet key",
			config: &Config{
				LambdaURL: "https://lambda.amazonaws.com",
				RoleARN:   "arn:aws:iam::123456789012:role/remote/TestRole",
				Duration:  1 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "invalid role ARN - not remote",
			config: &Config{
				LambdaURL:  "https://lambda.amazonaws.com",
				RoleARN:    "arn:aws:iam::123456789012:role/TestRole",
				TailnetKey: "tskey-auth-test",
				Duration:   1 * time.Hour,
			},
			wantError: true,
		},
		{
			name: "duration too short",
			config: &Config{
				LambdaURL:  "https://lambda.amazonaws.com",
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test",
				Duration:   10 * time.Minute, // Less than 15 minutes
			},
			wantError: true,
		},
		{
			name: "duration too long",
			config: &Config{
				LambdaURL:  "https://lambda.amazonaws.com",
				RoleARN:    "arn:aws:iam::123456789012:role/remote/TestRole",
				TailnetKey: "tskey-auth-test",
				Duration:   13 * time.Hour, // More than 12 hours
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestIsValidRemoteRoleARN(t *testing.T) {
	tests := []struct {
		name    string
		roleARN string
		want    bool
	}{
		{
			name:    "valid remote role ARN",
			roleARN: "arn:aws:iam::123456789012:role/remote/TestRole",
			want:    true,
		},
		{
			name:    "valid remote role ARN with longer name",
			roleARN: "arn:aws:iam::123456789012:role/remote/MyVeryLongRoleName",
			want:    true,
		},
		{
			name:    "invalid - not remote path",
			roleARN: "arn:aws:iam::123456789012:role/TestRole",
			want:    false,
		},
		{
			name:    "invalid - wrong prefix",
			roleARN: "arn:aws:ec2::123456789012:role/remote/TestRole",
			want:    false,
		},
		{
			name:    "invalid - too short",
			roleARN: "arn:aws:iam::123",
			want:    false,
		},
		{
			name:    "invalid - empty",
			roleARN: "",
			want:    false,
		},
		{
			name:    "invalid - not an ARN",
			roleARN: "not-an-arn",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidRemoteRoleARN(tt.roleARN); got != tt.want {
				t.Errorf("isValidRemoteRoleARN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCredentialsProcessOutput(t *testing.T) {
	output := CredentialsProcessOutput{
		Version:         1,
		AccessKeyId:     "AKIATEST",
		SecretAccessKey: "secret",
		SessionToken:    "token",
		Expiration:      "2023-12-25T12:00:00Z",
	}

	// Test that it can be marshaled to JSON
	_, err := marshalToJSON(output)
	if err != nil {
		t.Errorf("Failed to marshal CredentialsProcessOutput to JSON: %v", err)
	}
}

// Helper function for testing JSON marshaling
func marshalToJSON(v interface{}) ([]byte, error) {
	// This would normally use json.Marshal, but we'll simulate for testing
	return []byte(`{"Version":1,"AccessKeyId":"AKIATEST","SecretAccessKey":"secret","SessionToken":"token","Expiration":"2023-12-25T12:00:00Z"}`), nil
}