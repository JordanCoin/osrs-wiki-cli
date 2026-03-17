package dps

// Automatic Best-in-Slot gear finder.
// Scans equipment.json to find the highest offensive bonus items per slot.

import (
	"fmt"
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// FindBiSGear computes the best-in-slot gear set for a given combat style
// by scanning the equipment dataset for highest bonuses per slot.
func FindBiSGear(style string) (GearSet, error) {
	equipment, err := data.LoadEquipment()
	if err != nil {
		return GearSet{}, err
	}

	gear := GearSet{Style: style}

	switch style {
	case "melee":
		gear.Name = "Computed BiS Melee"
		gear = findBiSMelee(equipment, gear)
	case "ranged":
		gear.Name = "Computed BiS Ranged"
		gear = findBiSRanged(equipment, gear)
	case "magic":
		gear.Name = "Computed BiS Magic"
		gear = findBiSMagic(equipment, gear)
	default:
		return GearSet{}, fmt.Errorf("unknown style '%s' — use melee, ranged, or magic", style)
	}

	return gear, nil
}

func findBiSMelee(equipment []data.Equipment, gear GearSet) GearSet {
	bestPerSlot := map[string]struct {
		name  string
		score int
	}{}

	for _, e := range equipment {
		slot := strings.ToLower(e.Slot)
		if slot == "" || slot == "ammo" {
			continue
		}

		// Score: melee str is primary, slash attack is secondary
		score := e.Bonuses.Str*100 + e.Offensive.Slash

		// Skip items that are clearly not melee gear
		if slot == "weapon" {
			// Only consider melee weapons (has str or slash/stab/crush)
			if e.Bonuses.Str == 0 && e.Offensive.Slash == 0 && e.Offensive.Stab == 0 && e.Offensive.Crush == 0 {
				continue
			}
			// Skip ranged/magic weapons
			cat := strings.ToLower(e.Category)
			if strings.Contains(cat, "bow") || strings.Contains(cat, "crossbow") ||
				strings.Contains(cat, "staff") || strings.Contains(cat, "wand") ||
				strings.Contains(cat, "powered") || strings.Contains(cat, "thrown") ||
				strings.Contains(cat, "chinchompa") || strings.Contains(cat, "salamander") {
				continue
			}
		}

		// Skip degraded/broken variants
		name := e.Name
		if strings.Contains(name, "(broken)") || strings.Contains(name, "(damaged)") ||
			strings.Contains(name, "Deadman") || strings.Contains(name, "(bh)") ||
			strings.Contains(name, "(l)") {
			continue
		}

		if best, ok := bestPerSlot[slot]; !ok || score > best.score {
			bestPerSlot[slot] = struct {
				name  string
				score int
			}{name, score}
		}
	}

	// If best weapon is 2H, clear shield
	if best, ok := bestPerSlot["weapon"]; ok {
		for _, e := range equipment {
			if e.Name == best.name && e.IsTwoHanded {
				delete(bestPerSlot, "shield")
				break
			}
		}
	}

	// Map to gear set
	for slot, best := range bestPerSlot {
		switch slot {
		case "head":
			gear.Head = best.name
		case "body":
			gear.Body = best.name
		case "legs":
			gear.Legs = best.name
		case "hands":
			gear.Hands = best.name
		case "feet":
			gear.Feet = best.name
		case "cape":
			gear.Cape = best.name
		case "neck":
			gear.Neck = best.name
		case "ring":
			gear.Ring = best.name
		case "weapon":
			gear.Weapon = best.name
		case "shield":
			gear.Shield = best.name
		}
	}

	return gear
}

func findBiSRanged(equipment []data.Equipment, gear GearSet) GearSet {
	bestPerSlot := map[string]struct {
		name  string
		score int
	}{}

	for _, e := range equipment {
		slot := strings.ToLower(e.Slot)
		if slot == "" {
			continue
		}

		score := e.Bonuses.RangedStr*100 + e.Offensive.Ranged

		if slot == "weapon" {
			cat := strings.ToLower(e.Category)
			if !strings.Contains(cat, "bow") && !strings.Contains(cat, "crossbow") &&
				!strings.Contains(cat, "thrown") && !strings.Contains(cat, "chinchompa") &&
				!strings.Contains(cat, "salamander") && !strings.Contains(cat, "javelin") &&
				!strings.Contains(cat, "blowpipe") {
				continue
			}
		}

		name := e.Name
		if strings.Contains(name, "(broken)") || strings.Contains(name, "(damaged)") ||
			strings.Contains(name, "Deadman") || strings.Contains(name, "(bh)") {
			continue
		}

		if e.IsTwoHanded && slot == "weapon" {
			if best, ok := bestPerSlot["weapon"]; !ok || score > best.score {
				bestPerSlot["weapon"] = struct {
					name  string
					score int
				}{name, score}
				delete(bestPerSlot, "shield")
			}
			continue
		}

		if best, ok := bestPerSlot[slot]; !ok || score > best.score {
			bestPerSlot[slot] = struct {
				name  string
				score int
			}{name, score}
		}
	}

	for slot, best := range bestPerSlot {
		switch slot {
		case "head":
			gear.Head = best.name
		case "body":
			gear.Body = best.name
		case "legs":
			gear.Legs = best.name
		case "hands":
			gear.Hands = best.name
		case "feet":
			gear.Feet = best.name
		case "cape":
			gear.Cape = best.name
		case "neck":
			gear.Neck = best.name
		case "ring":
			gear.Ring = best.name
		case "ammo":
			gear.Ammo = best.name
		case "weapon":
			gear.Weapon = best.name
		case "shield":
			gear.Shield = best.name
		}
	}

	return gear
}

func findBiSMagic(equipment []data.Equipment, gear GearSet) GearSet {
	bestPerSlot := map[string]struct {
		name  string
		score int
	}{}

	for _, e := range equipment {
		slot := strings.ToLower(e.Slot)
		if slot == "" {
			continue
		}

		score := e.Bonuses.MagicStr*100 + e.Offensive.Magic

		if slot == "weapon" {
			cat := strings.ToLower(e.Category)
			if !strings.Contains(cat, "staff") && !strings.Contains(cat, "wand") &&
				!strings.Contains(cat, "powered") && !strings.Contains(cat, "salamander") {
				continue
			}
		}

		name := e.Name
		if strings.Contains(name, "(broken)") || strings.Contains(name, "(damaged)") ||
			strings.Contains(name, "Deadman") || strings.Contains(name, "(bh)") {
			continue
		}

		if e.IsTwoHanded && slot == "weapon" {
			if best, ok := bestPerSlot["weapon"]; !ok || score > best.score {
				bestPerSlot["weapon"] = struct {
					name  string
					score int
				}{name, score}
				delete(bestPerSlot, "shield")
			}
			continue
		}

		if best, ok := bestPerSlot[slot]; !ok || score > best.score {
			bestPerSlot[slot] = struct {
				name  string
				score int
			}{name, score}
		}
	}

	for slot, best := range bestPerSlot {
		switch slot {
		case "head":
			gear.Head = best.name
		case "body":
			gear.Body = best.name
		case "legs":
			gear.Legs = best.name
		case "hands":
			gear.Hands = best.name
		case "feet":
			gear.Feet = best.name
		case "cape":
			gear.Cape = best.name
		case "neck":
			gear.Neck = best.name
		case "ring":
			gear.Ring = best.name
		case "weapon":
			gear.Weapon = best.name
		case "shield":
			gear.Shield = best.name
		}
	}

	return gear
}
