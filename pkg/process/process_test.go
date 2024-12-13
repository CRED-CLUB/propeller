// pkg/process/process_test.go

package process

import (
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		wantOutput  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "echo command success",
			command:    "echo",
			args:       []string{"hello"},
			wantOutput: "hello\n",
			wantErr:    false,
		},
		{
			name:        "non-existent command",
			command:     "nonexistentcommand",
			args:        []string{},
			wantOutput:  "",
			wantErr:     true,
			errContains: "executable file not found",
		},
		{
			name:       "ls command success",
			command:    "ls",
			args:       []string{"-l"},
			wantOutput: "", // actual output will vary, we'll just check it's not empty
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Execute(tt.command, tt.args...)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Execute() error = %v, want error containing %v", err, tt.errContains)
					return
				}
			} else if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			// For commands like 'ls' where output varies, just check if output is present
			if tt.command == "ls" {
				if got == "" {
					t.Error("Execute() got empty output for ls command")
				}
				return
			}

			// Check output matches expected
			if got != tt.wantOutput {
				t.Errorf("Execute() got = %v, want %v", got, tt.wantOutput)
			}
		})
	}
}
