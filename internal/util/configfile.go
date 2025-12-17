package util

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"gopkg.in/yaml.v3"
)

type reverseWhoisConfig struct {
	APIKeys []string `yaml:"api_keys"`
}

func ResolveConfigPathNearExecutable(filename string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(exe)
	return filepath.Join(dir, filename), nil
}

func LoadOrInitAPIKeysYAML(path string) ([]string, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := reverseWhoisConfig{APIKeys: []string{}}
			b, mErr := yaml.Marshal(&cfg)
			if mErr != nil {
				return nil, fmt.Errorf("failed to marshal default yaml: %w", mErr)
			}
			if wErr := os.WriteFile(path, b, 0600); wErr != nil {
				return nil, fmt.Errorf("failed to create config file %q: %w", path, wErr)
			}
			return cfg.APIKeys, nil
		}
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg reverseWhoisConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml %q: %w", path, err)
	}

	out := make([]string, 0, len(cfg.APIKeys))
	for _, k := range cfg.APIKeys {
		if k == "" {
			continue
		}
		out = append(out, k)
	}
	return out, nil
}

type APIKeyProvider interface {
	Next() string
}

type staticKeyProvider struct {
	key string
}

func NewStaticKeyProvider(key string) APIKeyProvider {
	return &staticKeyProvider{key: key}
}

func (p *staticKeyProvider) Next() string {
	return p.key
}

type rotatingKeyProvider struct {
	keys []string
	i    uint64
}

func NewRotatingKeyProvider(keys []string) APIKeyProvider {
	return &rotatingKeyProvider{keys: keys}
}

func (p *rotatingKeyProvider) Next() string {
	if len(p.keys) == 0 {
		return ""
	}
	n := atomic.AddUint64(&p.i, 1) - 1
	return p.keys[int(n%uint64(len(p.keys)))]
}
