package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = flexibleString(sanitizeScalarText(str))
		return nil
	}

	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		*s = flexibleString(strconv.FormatFloat(num, 'f', -1, 64))
		return nil
	}

	return fmt.Errorf("unsupported scalar value %q", string(data))
}

// Monster represents a monster from the osrs-dps-calc dataset.
type Monster struct {
	ID      int            `json:"id"`
	Name    string         `json:"name"`
	Version string         `json:"version"`
	Image   string         `json:"image"`
	Level   int            `json:"level"`
	Speed   int            `json:"speed"`
	Size    int            `json:"size"`
	MaxHit  flexibleString `json:"max_hit"`
	Style   []string       `json:"style"`
	Skills  struct {
		Atk    int `json:"atk"`
		Def    int `json:"def"`
		HP     int `json:"hp"`
		Magic  int `json:"magic"`
		Ranged int `json:"ranged"`
		Str    int `json:"str"`
	} `json:"skills"`
	Offensive struct {
		Atk       int `json:"atk"`
		Magic     int `json:"magic"`
		MagicStr  int `json:"magic_str"`
		Ranged    int `json:"ranged"`
		RangedStr int `json:"ranged_str"`
		Str       int `json:"str"`
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
// Returns the best match, preferring exact display-name matches and sensible default versions.
func FindMonster(name string) (*Monster, error) {
	monsters, err := LoadMonsters()
	if err != nil {
		return nil, err
	}

	query := strings.TrimSpace(name)
	nameLower := strings.ToLower(query)

	for _, m := range monsters {
		if strings.EqualFold(m.DisplayName(), query) {
			match := m
			return &match, nil
		}
	}

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

	best := pickBestMonsterMatch(matches)
	return &best, nil
}

// FindMonsterVersion searches for a specific monster version.
func FindMonsterVersion(name, version string) (*Monster, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return FindMonster(name)
	}

	monsters, err := LoadMonsters()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	versionLower := strings.ToLower(version)
	var versions []string

	for _, m := range monsters {
		if strings.ToLower(m.Name) != nameLower {
			continue
		}

		versions = append(versions, m.Version)
		if strings.ToLower(m.Version) == versionLower {
			match := m
			return &match, nil
		}
	}

	if len(versions) == 0 {
		return FindMonster(name)
	}

	return nil, fmt.Errorf("monster '%s' version '%s' not found. Available versions: %s", name, version, strings.Join(versions, ", "))
}

func pickBestMonsterMatch(matches []Monster) Monster {
	best := matches[0]
	bestPriority := monsterVersionPriority(best.Version)

	for _, m := range matches[1:] {
		priority := monsterVersionPriority(m.Version)
		if priority < bestPriority {
			best = m
			bestPriority = priority
			continue
		}
		if priority == bestPriority && m.Skills.HP > best.Skills.HP {
			best = m
		}
	}

	return best
}

func monsterVersionPriority(version string) int {
	version = strings.ToLower(strings.TrimSpace(version))

	switch {
	case version == "":
		return 0
	case strings.Contains(version, "post-quest"), strings.Contains(version, "after quest"):
		return 1
	case strings.Contains(version, "normal"), strings.Contains(version, "regular"), strings.Contains(version, "standard"):
		return 2
	case strings.Contains(version, "quest"):
		return 4
	case strings.Contains(version, "awakened"),
		strings.Contains(version, "hard"),
		strings.Contains(version, "expert"),
		strings.Contains(version, "challenge"),
		strings.Contains(version, "entry"):
		return 5
	default:
		return 3
	}
}

func sanitizeScalarText(value string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return -1
		}
		if r == 127 {
			return -1
		}
		return r
	}, value)
}
