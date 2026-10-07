package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Ignore struct {
	Finding string `toml:"finding"`
	Object  string `toml:"object"`
	Reason  string `toml:"reason"`
}

// Config is the whole .kubot.toml. Deliberately small: suppression only.
type Config struct {
	Ignore []Ignore `toml:"ignore"`
}

// Load reads path ("" disables). Unknown credential-shaped keys are refused:
// config files get committed, secrets must never live in them.
func Load(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var blob map[string]any
	if err := toml.Unmarshal(raw, &blob); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	for k := range blob {
		switch k {
		case "ignore":
		default:
			if isCredentialKey(k) {
				return nil, fmt.Errorf("refusing credential-shaped key %q in config (secrets don't belong in committed files)", k)
			}
		}
	}
	var c Config
	if err := toml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	for i, r := range c.Ignore {
		if r.Finding == "" {
			return nil, fmt.Errorf("ignore rule %d: finding is required", i)
		}
		if r.Reason == "" {
			return nil, fmt.Errorf("ignore rule %d: reason is required (write why, future you will ask)", i)
		}
	}
	return &c, nil
}

func isCredentialKey(k string) bool {
	lower := strings.ToLower(k)
	for _, sub := range []string{"password", "secret", "token", "key", "dsn", "url", "connection"} {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// Discover returns --config, $KUBOT_CONFIG, or ./.kubot.toml, else "".
func Discover(flag, env string) string {
	if flag != "" {
		return flag
	}
	if env != "" {
		return env
	}
	local := filepath.Join(".kubot.toml")
	if _, err := os.Stat(local); err == nil {
		return local
	}
	return ""
}

// Match reports whether any ignore rule silences a finding.
func (c *Config) Match(finding, namespace, resource string) (string, bool) {
	joined := namespace + "/" + resource
	base := resource
	if i := strings.Index(resource, "/"); i >= 0 {
		base = resource[i+1:]
	}
	candidates := []string{joined, resource, namespace + "/" + base}
	for _, r := range c.Ignore {
		if r.Finding != finding {
			continue
		}
		if r.Object == "" || r.Object == "*" {
			return r.Reason, true
		}
		if strings.HasSuffix(r.Object, "/*") && strings.HasPrefix(joined, r.Object[:len(r.Object)-1]) {
			return r.Reason, true
		}
		for _, cand := range candidates {
			if ok, _ := path.Match(r.Object, cand); ok {
				return r.Reason, true
			}
		}
	}
	return "", false
}
