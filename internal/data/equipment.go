package data

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Equipment represents a piece of equipment from the osrs-dps-calc dataset.
type Equipment struct {
	Name       string `json:"name"`
	ID         int    `json:"id"`
	Version    string `json:"version"`
	Slot       string `json:"slot"`
	Image      string `json:"image"`
	Speed      int    `json:"speed"`
	Category   string `json:"category"`
	IsTwoHanded bool  `json:"isTwoHanded"`
	Bonuses    struct {
		Str       int `json:"str"`
		RangedStr int `json:"ranged_str"`
		MagicStr  int `json:"magic_str"`
		Prayer    int `json:"prayer"`
	} `json:"bonuses"`
	Offensive struct {
		Stab   int `json:"stab"`
		Slash  int `json:"slash"`
		Crush  int `json:"crush"`
		Magic  int `json:"magic"`
		Ranged int `json:"ranged"`
	} `json:"offensive"`
	Defensive struct {
		Stab   int `json:"stab"`
		Slash  int `json:"slash"`
		Crush  int `json:"crush"`
		Magic  int `json:"magic"`
		Ranged int `json:"ranged"`
	} `json:"defensive"`
}

var cachedEquipment []Equipment

// LoadEquipment loads and caches the equipment dataset.
func LoadEquipment() ([]Equipment, error) {
	if cachedEquipment != nil {
		return cachedEquipment, nil
	}
	path, err := EnsureEquipment()
	if err != nil {
		return nil, err
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read equipment.json: %w", err)
	}
	if err := json.Unmarshal(f, &cachedEquipment); err != nil {
		return nil, fmt.Errorf("cannot parse equipment.json: %w", err)
	}
	return cachedEquipment, nil
}

// FindEquipment searches for equipment by name (case-insensitive, exact match).
func FindEquipment(name string) (*Equipment, error) {
	equipment, err := LoadEquipment()
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	for _, e := range equipment {
		if strings.ToLower(e.Name) == nameLower {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("equipment '%s' not found", name)
}

// FindEquipmentBySlot returns all equipment for a given slot.
func FindEquipmentBySlot(slot string) ([]Equipment, error) {
	equipment, err := LoadEquipment()
	if err != nil {
		return nil, err
	}

	var results []Equipment
	slotLower := strings.ToLower(slot)
	for _, e := range equipment {
		if strings.ToLower(e.Slot) == slotLower {
			results = append(results, e)
		}
	}
	return results, nil
}
