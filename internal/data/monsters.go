package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Monster represents a monster from the osrs-dps-calc dataset.
type Monster struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Image   string `json:"image"`
	Level   int    `json:"level"`
	Speed   int    `json:"speed"`
	Size    int    `json:"size"`
	MaxHit  string `json:"max_hit"`
	Style   []string `json:"style"`
	Skills  struct {
		Atk    int `json:"atk"`
		Def    int `json:"def"`
		HP     int `json:"hp"`
		Magic  int `json:"magic"`
		Ranged int `json:"ranged"`
		Str    int `json:"str"`
	} `json:"skills"`
	Offensive struct {
		Atk      int `json:"atk"`
		Magic    int `json:"magic"`
		MagicStr int `json:"magic_str"`
		Ranged   int `json:"ranged"`
		RangedStr int `json:"ranged_str"`
		Str      int `json:"str"`
	} `json:"offensive"`
	Defensive struct {
		FlatArmour int `json:"flat_armour"`
		Crush      int `json:"crush"`
		Magic      int `json:"magic"`
		Heavy      int `json:"heavy"`
		Standard   int `json:"standard"`
		Light      int `json:"light"`
		Slash      int `json:"slash"`
		Stab       int `json:"stab"`
	} `json:"defensive"`
	Attributes []string `json:"attributes"`
	Weakness   *struct {
		Element  string `json:"element"`
		Severity int    `json:"severity"`
	} `json:"weakness"`
}

// DisplayName returns the monster name with version if applicable.
func (m *Monster) DisplayName() string {
	if m.Version != "" {
		return fmt.Sprintf("%s (%s)", m.Name, m.Version)
	}
	return m.Name
}

var cachedMonsters []Monster

// LoadMonsters loads and caches the monster dataset.
func LoadMonsters() ([]Monster, error) {
	if cachedMonsters != nil {
		return cachedMonsters, nil
	}
	path, err := EnsureMonsters()
	if err != nil {
		return nil, err
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read monsters.json: %w", err)
	}
	if err := json.Unmarshal(f, &cachedMonsters); err != nil {
		return nil, fmt.Errorf("cannot parse monsters.json: %w", err)
	}
	return cachedMonsters, nil
}

// FindMonster searches for a monster by name (case-insensitive).
// Returns the best match, preferring exact matches then the highest-level version.
func FindMonster(name string) (*Monster, error) {
	monsters, err := LoadMonsters()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	var matches []Monster

	for _, m := range monsters {
		if strings.ToLower(m.Name) == nameLower {
			matches = append(matches, m)
		}
	}

	if len(matches) == 0 {
		// Try partial match
		for _, m := range monsters {
			if strings.Contains(strings.ToLower(m.Name), nameLower) {
				matches = append(matches, m)
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("monster '%s' not found. Check the spelling", name)
	}

	// Prefer the version with highest HP (usually the "real" version)
	best := matches[0]
	for _, m := range matches[1:] {
		if m.Skills.HP > best.Skills.HP {
			best = m
		}
	}
	return &best, nil
}

// FindMonsterVersion searches for a specific monster version.
func FindMonsterVersion(name, version string) (*Monster, error) {
	monsters, err := LoadMonsters()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)

	for _, m := range monsters {
		if strings.ToLower(m.Name) == nameLower && strings.ToLower(m.Version) == versionLower {
			return &m, nil
		}
	}
	return FindMonster(name)
}
