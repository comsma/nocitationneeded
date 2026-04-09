package config

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	HCaptcha HCaptcha `yaml:"hcaptcha"`
}

// LoadConfig returns a Config struct from a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	var cfg Config
	var k = koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("loading config from file: %w", err)
	}

	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
