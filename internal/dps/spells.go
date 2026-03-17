package dps

// Spell max hits and related calculations.
// Ported from weirdgloop/osrs-dps-calc types/Spell.ts + PlayerVsNPCCalc.ts

// SpellMaxHit returns the base max hit for a spell by name.
func SpellMaxHit(spellName string, magicLevel int) int {
	switch spellName {
	// Standard spellbook — Strike
	case "Wind Strike":
		return 2
	case "Water Strike":
		return 4
	case "Earth Strike":
		return 6
	case "Fire Strike":
		return 8
	// Standard spellbook — Bolt
	case "Wind Bolt":
		return 9
	case "Water Bolt":
		return 10
	case "Earth Bolt":
		return 11
	case "Fire Bolt":
		return 12
	// Standard spellbook — Blast
	case "Wind Blast":
		return 13
	case "Water Blast":
		return 14
	case "Earth Blast":
		return 15
	case "Fire Blast":
		return 16
	// Standard spellbook — Wave
	case "Wind Wave":
		return 17
	case "Water Wave":
		return 18
	case "Earth Wave":
		return 19
	case "Fire Wave":
		return 20
	// Standard spellbook — Surge
	case "Wind Surge":
		return 21
	case "Water Surge":
		return 22
	case "Earth Surge":
		return 23
	case "Fire Surge":
		return 24

	// God spells
	case "Saradomin Strike":
		return 20
	case "Claws of Guthix":
		return 20
	case "Flames of Zamorak":
		return 20

	// Ancient spellbook — Rush
	case "Smoke Rush":
		return 13
	case "Shadow Rush":
		return 14
	case "Blood Rush":
		return 15
	case "Ice Rush":
		return 16
	// Ancient spellbook — Burst
	case "Smoke Burst":
		return 17
	case "Shadow Burst":
		return 18
	case "Blood Burst":
		return 21
	case "Ice Burst":
		return 22
	// Ancient spellbook — Blitz
	case "Smoke Blitz":
		return 23
	case "Shadow Blitz":
		return 24
	case "Blood Blitz":
		return 25
	case "Ice Blitz":
		return 26
	// Ancient spellbook — Barrage
	case "Smoke Barrage":
		return 27
	case "Shadow Barrage":
		return 28
	case "Blood Barrage":
		return 29
	case "Ice Barrage":
		return 30

	// Arceuus spellbook
	case "Ghostly Grasp":
		return 12
	case "Skeletal Grasp":
		return 17
	case "Undead Grasp":
		return 24
	case "Inferior Demonbane":
		return 16
	case "Superior Demonbane":
		return 23
	case "Dark Demonbane":
		return 30

	// Special spells
	case "Crumble Undead":
		return 15
	case "Iban Blast":
		return 25
	case "Magic Dart":
		return magicDartMaxHit(magicLevel, false)

	// Bind spells (0 damage)
	case "Bind", "Snare", "Entangle":
		return 0

	default:
		return 0
	}
}

func magicDartMaxHit(magicLevel int, slayerStaffE bool) int {
	if slayerStaffE {
		return 13 + magicLevel/6
	}
	return 10 + magicLevel/10
}

// SpellElement returns the element of a spell for weakness matching.
func SpellElement(spellName string) string {
	switch {
	case contains(spellName, "Fire", "Flames of Zamorak"):
		return "fire"
	case contains(spellName, "Water"):
		return "water"
	case contains(spellName, "Earth"):
		return "earth"
	case contains(spellName, "Wind"):
		return "air"
	case contains(spellName, "Ice"):
		return "water"
	case contains(spellName, "Smoke"):
		return "fire"
	case contains(spellName, "Shadow"):
		return "earth"
	case contains(spellName, "Blood"):
		return "fire"
	default:
		return ""
	}
}

// IsBoltSpell returns true for bolt-tier spells (for chaos gauntlets).
func IsBoltSpell(name string) bool {
	return contains(name, "Bolt")
}

// IsStandardSpellbook returns true for standard spellbook spells.
func IsStandardSpellbook(name string) bool {
	standards := []string{
		"Strike", "Bolt", "Blast", "Wave", "Surge",
		"Bind", "Snare", "Entangle",
		"Crumble Undead", "Iban Blast", "Magic Dart",
		"Saradomin Strike", "Claws of Guthix", "Flames of Zamorak",
	}
	for _, s := range standards {
		if contains(name, s) {
			return true
		}
	}
	return false
}

// IsDemonbaneSpell returns true for demonbane spells.
func IsDemonbaneSpell(name string) bool {
	return contains(name, "Demonbane")
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
