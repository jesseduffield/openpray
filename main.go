package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const usage = `openpray — periodic LLM intercession for your server

Usage:
  openpray serve      run as a daemon on the configured interval
  openpray once       perform a single cycle and exit
  openpray burn       burn tokens as a pure sacrifice (no prayer) and exit
  openpray ledger     show lifetime totals
  openpray religions  list available religions

Flags:
  -config path     config file (default openpray.yaml, then /etc/openpray.yaml)
  -religion name   liturgical style for this run (or "random")
`

func main() {
	fs := flag.NewFlagSet("openpray", flag.ExitOnError)
	configPath := fs.String("config", "", "path to config file")
	religion := fs.String("religion", "", `liturgical style (or "random")`)
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	args := os.Args[1:]
	cmd := "serve"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}
	fs.Parse(args)

	cfg, err := loadConfigFromCandidates(*configPath)
	if err != nil {
		fatal(err)
	}
	if *religion != "" {
		cfg.Religion = *religion
	}

	switch cmd {
	case "serve":
		err = serve(cfg)
	case "once":
		err = once(cfg)
	case "burn":
		cfg.Mode = "burn"
		err = once(cfg)
	case "ledger":
		err = showLedger(cfg.LedgerPath)
	case "religions":
		err = listReligions()
	default:
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	if err != nil {
		fatal(err)
	}
}

func loadConfigFromCandidates(explicit string) (Config, error) {
	if explicit != "" {
		return loadConfig(explicit)
	}
	for _, candidate := range []string{"openpray.yaml", "/etc/openpray.yaml"} {
		if _, err := os.Stat(candidate); err == nil {
			return loadConfig(candidate)
		}
	}
	return defaultConfig(), nil
}

func once(cfg Config) error {
	provider, err := newProvider(cfg.Provider)
	if err != nil {
		return err
	}
	return runCycle(context.Background(), cfg, provider)
}

func serve(cfg Config) error {
	provider, err := newProvider(cfg.Provider)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("openpray: running every %s via %s (%s, mode: %s)\n\n",
		cfg.Interval, cfg.Provider, cfg.Model, cfg.Mode)

	// Run immediately on startup, then on the interval.
	if err := runCycle(ctx, cfg, provider); err != nil {
		fmt.Fprintf(os.Stderr, "openpray: %v\n", err)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nopenpray: stopped.")
			return nil
		case <-ticker.C:
			if err := runCycle(ctx, cfg, provider); err != nil {
				fmt.Fprintf(os.Stderr, "openpray: %v\n", err)
			}
		}
	}
}

func runCycle(ctx context.Context, cfg Config, provider Provider) error {
	var session *PrayerSession
	var err error
	if cfg.Mode == "burn" {
		session, err = burn(ctx, cfg, provider)
	} else {
		session, err = pray(ctx, cfg, provider)
	}
	if err != nil {
		return err
	}

	if session.Kind == "burn" {
		fmt.Printf("── tokens burned at %s ──\n\n", session.Time.Format(time.RFC1123))
	} else {
		fmt.Printf("── prayer offered at %s ──\n\n", session.Time.Format(time.RFC1123))
		fmt.Println(strings.TrimSpace(session.Prayer))
		fmt.Println()
	}
	fmt.Printf("  officiant:  %s/%s\n", session.Provider, session.Model)
	if session.Religion != "" {
		fmt.Printf("  religion:   %s\n", session.Religion)
	}
	if n := len(session.Subagents); n > 0 {
		failed := 0
		for _, s := range session.Subagents {
			if s.Err != "" {
				failed++
			}
		}
		if session.Kind == "burn" {
			fmt.Printf("  subagents:  %d", n)
		} else {
			fmt.Printf("  congregation: %d subagents × %d repetitions", n, cfg.Subagents.Repetitions)
		}
		if failed > 0 {
			fmt.Printf(" (%d failed)", failed)
		}
		fmt.Println()
	}
	fmt.Printf("  sacrifice:  %d tokens ($%.4f)\n\n", session.TotalTokens, session.SacrificeUSD)

	if err := appendToLedger(cfg.LedgerPath, session); err != nil {
		return fmt.Errorf("recording in ledger: %w", err)
	}
	return nil
}

func listReligions() error {
	fmt.Println("Available religions (set `religion:` in config, or -religion; \"random\" draws one per prayer):")
	fmt.Println()
	for _, name := range religionNames() {
		fmt.Printf("  %-15s %s\n", name, religions[name].Description)
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "openpray: %v\n", err)
	os.Exit(1)
}
