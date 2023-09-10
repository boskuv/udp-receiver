package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct for webapp config
type Config struct {
	Services         map[string]int `yaml:"services"`
	SleepTimeSec     int            `yaml:"sleepTimeSec"`
	AnswerTimeoutSec int            `yaml:"answerTimeoutSec"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*Config, error) {
	// Create config structure
	config := &Config{map[string]int{"test_service": 8789}, 1, 5}

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
	flag.StringVar(&configPath, "config", "./config.yml", "path to config file") // TODO: get path from root directory

	// Actually parse the flags
	flag.Parse()

	// Validate the path first
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}

	// Return the configuration path
	return configPath, nil
}
