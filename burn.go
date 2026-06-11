package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

const burnSystemPrompt = `You are a token furnace. Your output is discarded
unread. Do not produce anything meaningful.`

const burnPrompt = `Emit the word "burn", repeated, separated by spaces, with
no other content, until your output budget is exhausted.`

// burn performs one pure-sacrifice cycle: tokens are generated and
// discarded. No prayer is composed. If subagents are enabled, each subagent
// burns the same budget concurrently.
func burn(ctx context.Context, cfg Config, provider Provider) (*PrayerSession, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	completion, err := provider.Complete(ctx, cfg.Model, burnSystemPrompt, burnPrompt, cfg.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("burning tokens: %w", err)
	}

	session := &PrayerSession{
		Time:         time.Now(),
		Kind:         "burn",
		Hostname:     hostname,
		Provider:     provider.Name(),
		Model:        cfg.Model,
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

		results := make([]SubagentResult, cfg.Subagents.Count)
		var wg sync.WaitGroup
		for i := 0; i < cfg.Subagents.Count; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				results[i] = SubagentResult{Index: i}
				c, err := provider.Complete(ctx, subModel, burnSystemPrompt, burnPrompt, cfg.MaxTokens)
				if err != nil {
					results[i].Err = err.Error()
					return
				}
				results[i].InputTokens = c.InputTokens
				results[i].OutputTokens = c.OutputTokens
			}(i)
		}
		wg.Wait()

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
