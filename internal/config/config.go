package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	HCaptcha HCaptcha `koanf:"hcaptcha"`
}

// LoadConfig loads configuration from a YAML file then overlays APP_* environment variables.
// The config file is optional; if absent, env vars alone are used.
// Env var mapping: APP_HCAPTCHA_SITEKEY -> hcaptcha.sitekey
func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("loading config from file: %w", err)
		}
	}

	if err := k.Load(env.Provider("APP_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "APP_")), "_", ".")
	}), nil); err != nil {
		return nil, fmt.Errorf("loading config from env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	return &cfg, nil
}
