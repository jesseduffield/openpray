package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Pricing is the cost of a model in dollars per million tokens.
type Pricing struct {
	InputPerMTok  float64 `yaml:"input_per_mtok"`
	OutputPerMTok float64 `yaml:"output_per_mtok"`
}

type SubagentConfig struct {
	Enabled     bool `yaml:"enabled"`
	Count       int  `yaml:"count"`
	Repetitions int  `yaml:"repetitions"`
	// Model lets the congregation chant on a cheaper model than the
	// orchestrator. Empty means same model as the orchestrator.
	Model string `yaml:"model"`
}

type Config struct {
	Provider string `yaml:"provider"` // anthropic | openai | gemini
	Model    string `yaml:"model"`
	// Mode is "prayer" (compose and log a prayer) or "burn" (generate and
	// discard tokens with no prayer; only the expenditure is recorded).
	Mode      string        `yaml:"mode"`
	Interval  time.Duration `yaml:"interval"`
	MaxTokens int64         `yaml:"max_tokens"`
	// Religion selects the liturgical style, or "random" to draw a fresh
	// denomination each prayer cycle. Run `openpray religions` to list.
	Religion     string         `yaml:"religion"`
	PrayerPrompt string         `yaml:"prayer_prompt"` // optional override
	Subagents    SubagentConfig `yaml:"subagents"`
	LedgerPath   string         `yaml:"ledger_path"`
	// PricingOverride lets you declare the sacrifice value of models the
	// built-in registry doesn't know about (or correct stale prices).
	PricingOverride map[string]Pricing `yaml:"pricing"`
}

func defaultConfig() Config {
	return Config{
		Provider:  "anthropic",
		Model:     "claude-opus-4-8",
		Mode:      "prayer",
		Interval:  time.Hour,
		MaxTokens: 1024,
		Religion:  "random",
		Subagents: SubagentConfig{
			Enabled:     false,
			Count:       3,
			Repetitions: 10,
		},
		LedgerPath: "~/.openpray/ledger.jsonl",
	}
}

func loadConfig(path string) (Config, error) {
	cfg := defaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", path, err)
	}
	if cfg.Subagents.Count < 1 {
		cfg.Subagents.Count = 1
	}
	if cfg.Subagents.Repetitions < 1 {
		cfg.Subagents.Repetitions = 10
	}
	if cfg.MaxTokens < 1 {
		cfg.MaxTokens = 1024
	}
	if cfg.Mode == "" {
		cfg.Mode = "prayer"
	}
	if cfg.Mode != "prayer" && cfg.Mode != "burn" {
		return cfg, fmt.Errorf("unknown mode %q (want prayer or burn)", cfg.Mode)
	}
	return cfg, nil
}

func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
