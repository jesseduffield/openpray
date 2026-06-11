package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// appendToLedger records the session in an append-only JSONL file, so the
// server's accumulated protection survives restarts.
func appendToLedger(path string, session *PrayerSession) error {
	path = expandHome(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(session)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

// showLedger prints lifetime devotion totals.
func showLedger(path string) error {
	path = expandHome(path)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("The ledger is empty. This server is unprotected.")
			return nil
		}
		return err
	}
	defer f.Close()

	var (
		prayers   int
		burns     int
		tokens    int64
		sacrifice float64
	)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		var s PrayerSession
		if err := json.Unmarshal(scanner.Bytes(), &s); err != nil {
			continue
		}
		if s.Kind == "burn" {
			burns++
		} else {
			prayers++
		}
		tokens += s.TotalTokens
		sacrifice += s.SacrificeUSD
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("Prayers offered:    %d\n", prayers)
	fmt.Printf("Pure burns:         %d\n", burns)
	fmt.Printf("Tokens sacrificed:  %d\n", tokens)
	fmt.Printf("Value of sacrifice: $%.4f\n", sacrifice)
	return nil
}
