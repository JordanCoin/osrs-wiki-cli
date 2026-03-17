package data

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	monstersURL  = "https://raw.githubusercontent.com/weirdgloop/osrs-dps-calc/main/cdn/json/monsters.json"
	equipmentURL = "https://raw.githubusercontent.com/weirdgloop/osrs-dps-calc/main/cdn/json/equipment.json"
	maxAge       = 7 * 24 * time.Hour // re-download weekly
)

// CacheDir returns the data cache directory, creating it if needed.
func CacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".osrs-wiki", "data")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create cache dir: %w", err)
	}
	return dir, nil
}

// EnsureFile downloads a file if it's missing or stale.
func EnsureFile(name, url string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)

	info, err := os.Stat(path)
	if err == nil && time.Since(info.ModTime()) < maxAge {
		return path, nil // cached and fresh
	}

	fmt.Fprintf(os.Stderr, "Downloading %s...\n", name)
	client := &http.Client{Timeout: 60 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "osrs-wiki-cli (github.com/JordanCoin/osrs-wiki-cli)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("cannot write cache file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("download incomplete: %w", err)
	}

	return path, nil
}

// EnsureMonsters returns the path to the cached monsters.json.
func EnsureMonsters() (string, error) {
	return EnsureFile("monsters.json", monstersURL)
}

// EnsureEquipment returns the path to the cached equipment.json.
func EnsureEquipment() (string, error) {
	return EnsureFile("equipment.json", equipmentURL)
}
