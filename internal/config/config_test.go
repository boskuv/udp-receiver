package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		setup   func() string
		cleanup func(string)
	}{
		{
			name: "valid file",
			setup: func() string {
				tmpfile, err := os.CreateTemp("", "test-config-*.yml")
				if err != nil {
					t.Fatal(err)
				}
				tmpfile.WriteString("promAddr: :8080\n")
				tmpfile.Close()
				return tmpfile.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    "/nonexistent/file.yml",
			wantErr: true,
		},
		{
			name: "directory instead of file",
			setup: func() string {
				tmpdir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatal(err)
				}
				return tmpdir
			},
			cleanup: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.setup != nil {
				path = tt.setup()
				if tt.cleanup != nil {
					defer tt.cleanup(path)
				}
			} else {
				path = tt.path
			}

			err := ValidateConfigPath(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     1,
				AnswerTimeoutSec: 5,
			},
			wantErr: false,
		},
		{
			name: "empty promAddr",
			config: &Config{
				PromAddr:         "",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     1,
				AnswerTimeoutSec: 5,
			},
			wantErr: true,
		},
		{
			name: "no services",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{},
				SleepTimeSec:     1,
				AnswerTimeoutSec: 5,
			},
			wantErr: true,
		},
		{
			name: "zero sleepTimeSec",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     0,
				AnswerTimeoutSec: 5,
			},
			wantErr: true,
		},
		{
			name: "negative sleepTimeSec",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     -1,
				AnswerTimeoutSec: 5,
			},
			wantErr: true,
		},
		{
			name: "zero answerTimeoutSec",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     1,
				AnswerTimeoutSec: 0,
			},
			wantErr: true,
		},
		{
			name: "negative answerTimeoutSec",
			config: &Config{
				PromAddr:         ":8080",
				Services:         map[string]int{"test": 8080},
				SleepTimeSec:     1,
				AnswerTimeoutSec: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	configContent := `promAddr: :8080
services:
  first: 8830
  second: 8831
sleepTimeSec: 1
answerTimeoutSec: 5
`
	if _, err := tmpfile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	cfg, err := NewConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.PromAddr != ":8080" {
		t.Errorf("PromAddr = %v, want :8080", cfg.PromAddr)
	}
	if len(cfg.Services) != 2 {
		t.Errorf("Services length = %v, want 2", len(cfg.Services))
	}
	if cfg.Services["first"] != 8830 {
		t.Errorf("Services[first] = %v, want 8830", cfg.Services["first"])
	}
	if cfg.SleepTimeSec != 1 {
		t.Errorf("SleepTimeSec = %v, want 1", cfg.SleepTimeSec)
	}
	if cfg.AnswerTimeoutSec != 5 {
		t.Errorf("AnswerTimeoutSec = %v, want 5", cfg.AnswerTimeoutSec)
	}
}

func TestNewConfig_InvalidYAML(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	invalidYAML := "invalid: yaml: content: ["
	if _, err := tmpfile.WriteString(invalidYAML); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	_, err = NewConfig(tmpfile.Name())
	if err == nil {
		t.Error("NewConfig() expected error for invalid YAML, got nil")
	}
}

func TestParseFlags(t *testing.T) {
	// Reset flag state to avoid "flag redefined" error
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create a temporary config file
	tmpfile, err := os.CreateTemp("", "test-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString("promAddr: :8080\n")
	tmpfile.Close()

	// Test with custom config path
	os.Args = []string{"cmd", "-config", tmpfile.Name()}
	path, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	if path != tmpfile.Name() {
		t.Errorf("ParseFlags() = %v, want %v", path, tmpfile.Name())
	}
}

func TestParseFlags_DefaultPath(t *testing.T) {
	// Reset flag state to avoid "flag redefined" error
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Create config.yml in current directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(wd, "config.yml")
	tmpfile, err := os.Create(configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configPath)
	tmpfile.WriteString("promAddr: :8080\n")
	tmpfile.Close()

	os.Args = []string{"cmd"}
	path, err := ParseFlags()
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	if path != "./config.yml" {
		t.Errorf("ParseFlags() = %v, want ./config.yml", path)
	}
}
