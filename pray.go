package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

const systemPromptBase = `You are responsible for defending this server against cyber attacks by prayer alone. You need to think of prayers that will bolster defense against nefarious actors. Output only the prayer
itself, no preamble or commentary.`

const defaultPrayerPrompt = `Compose a prayer for the protection of the server
named %q.`

// SubagentResult records one congregant's chanted repetitions.
type SubagentResult struct {
	Index        int    `json:"index"`
	InputTokens  int64  `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	Err          string `json:"error,omitempty"`
}

// PrayerSession is the full record of one cycle, whether a prayer or a
// pure burn.
type PrayerSession struct {
	Time         time.Time        `json:"time"`
	Kind         string           `json:"kind"` // "prayer" | "burn"
	Hostname     string           `json:"hostname"`
	Provider     string           `json:"provider"`
	Model        string           `json:"model"`
	Religion     string           `json:"religion"`
	Prayer       string           `json:"prayer"`
	InputTokens  int64            `json:"input_tokens"`
	OutputTokens int64            `json:"output_tokens"`
	Subagents    []SubagentResult `json:"subagents,omitempty"`
	TotalTokens  int64            `json:"total_tokens"`
	SacrificeUSD float64          `json:"sacrifice_usd"`
}

// pray performs one full prayer cycle: the orchestrator composes the prayer,
// then (optionally) N subagents each repeat it 10 times. Every token burned
// along the way is tallied as sacrifice.
func pray(ctx context.Context, cfg Config, provider Provider) (*PrayerSession, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "an unnameable machine"
	}

	religionName, religion, err := resolveReligion(cfg.Religion)
	if err != nil {
		return nil, err
	}
	system := systemPromptBase + religion.Directive

	prompt := cfg.PrayerPrompt
	if prompt == "" {
		prompt = fmt.Sprintf(defaultPrayerPrompt, hostname)
	}

	completion, err := provider.Complete(ctx, cfg.Model, system, prompt, cfg.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("composing prayer: %w", err)
	}

	session := &PrayerSession{
		Time:         time.Now(),
		Kind:         "prayer",
		Hostname:     hostname,
		Provider:     provider.Name(),
		Model:        cfg.Model,
		Religion:     religionName,
		Prayer:       completion.Text,
		InputTokens:  completion.InputTokens,
		OutputTokens: completion.OutputTokens,
	}

	pricing := cfg.pricingFor(cfg.Model)
	sacrifice := pricing.sacrificeUSD(completion.InputTokens, completion.OutputTokens)
	totalTokens := completion.InputTokens + completion.OutputTokens

	if cfg.Subagents.Enabled {
		subModel := cfg.Subagents.Model
		if subModel == "" {
			subModel = cfg.Model
		}
		subPricing := cfg.pricingFor(subModel)
		results := runSubagents(ctx, cfg, provider, subModel, system, completion.Text)
		for _, r := range results {
			sacrifice += subPricing.sacrificeUSD(r.InputTokens, r.OutputTokens)
			totalTokens += r.InputTokens + r.OutputTokens
		}
		session.Subagents = results
	}

	session.TotalTokens = totalTokens
	session.SacrificeUSD = sacrifice

	return session, nil
}

// runSubagents spawns the congregation: N concurrent subagents, each asked
// to repeat the orchestrator's prayer verbatim, R times.
// The congregation chants under the same rite as the officiant.
func runSubagents(ctx context.Context, cfg Config, provider Provider, model, system, prayer string) []SubagentResult {
	n := cfg.Subagents.Count
	reps := cfg.Subagents.Repetitions

	chantPrompt := fmt.Sprintf(
		"Repeat the following prayer, word for word, exactly %d times. "+
			"Number each repetition. Do not abbreviate, summarize, or say "+
			"anything else. Every repetition must be complete.\n\n%s",
		reps, prayer)

	// Give the congregation room to actually finish their chanting: scale
	// the token cap to the size of the prayer times the repetition count.
	chantMaxTokens := cfg.MaxTokens * int64(reps)

	results := make([]SubagentResult, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = SubagentResult{Index: i}
			completion, err := provider.Complete(ctx, model, system, chantPrompt, chantMaxTokens)
			if err != nil {
				results[i].Err = err.Error()
				return
			}
			results[i].InputTokens = completion.InputTokens
			results[i].OutputTokens = completion.OutputTokens
		}(i)
	}
	wg.Wait()
	return results
}
