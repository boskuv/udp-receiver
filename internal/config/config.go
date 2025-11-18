package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	PromAddr         string         `yaml:"promAddr"`
	Services         map[string]int `yaml:"services"`
	SleepTimeSec     int            `yaml:"sleepTimeSec"`
	AnswerTimeoutSec int            `yaml:"answerTimeoutSec"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.PromAddr == "" {
		return fmt.Errorf("promAddr is required")
	}
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}
	if c.SleepTimeSec <= 0 {
		return fmt.Errorf("sleepTimeSec must be greater than 0")
	}
	if c.AnswerTimeoutSec <= 0 {
		return fmt.Errorf("answerTimeoutSec must be greater than 0")
	}
	return nil
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("os.Open(%q): %w", configPath, err)
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("yaml.NewDecoder(..).Decode(..): %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// ValidateConfigPath just makes sure, that the path provided is a file,
// that can be read
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ValidateConfigPath(%s): %w", path, err)
	}
	if s.IsDir() {
		var ErrIsDir = errors.New("it's a directory, not a file")
		return fmt.Errorf("%w: %s", ErrIsDir, path)
	}
	return nil
}

// ParseFlags will create and parse the CLI flags
// and return the path to be used elsewhere
func ParseFlags() (string, error) {
	// String that contains the configured configuration path
	var configPath string

	// Set up a CLI flag called "-config" to allow users
	// to supply the configuration file
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}
