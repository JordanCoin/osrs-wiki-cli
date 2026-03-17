package dps

// Special weapon effects and modifiers.
// Ported from weirdgloop/osrs-dps-calc PlayerVsNPCCalc.ts

import (
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// ── Scythe of vitur ────────────────────────────────────────────────

// ScytheExpectedMax computes the effective max hit for scythe,
// accounting for multi-hitsplats on large monsters.
// Hit 0: max, Hit 1: max/2, Hit 2: max/4
func ScytheExpectedMax(baseMax int, monsterSize int) int {
	hits := monsterSize
	if hits < 1 {
		hits = 1
	}
	if hits > 3 {
		hits = 3
	}

	total := 0
	for i := 0; i < hits; i++ {
		hitMax := baseMax
		for j := 0; j < i; j++ {
			hitMax = hitMax / 2
		}
		total += hitMax
	}
	return total
}

// ── Dharok's set effect ────────────────────────────────────────────

// DharokMaxHit applies the Dharok's set effect multiplier.
// Factor: (10000 + (maxHP - currentHP) * maxHP) / 10000
func DharokMaxHit(baseMax, maxHP, currentHP int) int {
	return baseMax * (10000 + (maxHP-currentHP)*maxHP) / 10000
}

// ── Verac's set effect ─────────────────────────────────────────────

// VeracExpectedDPS computes expected DPS with Verac's set effect.
// 75% normal distribution, 25% ignores defence (100% accuracy, +1 max).
func VeracExpectedDPS(accuracy float64, maxHit int, atkSpeed int) float64 {
	normalDPS := accuracy * float64(maxHit) / 2.0 / (float64(atkSpeed) * SecondsPerTick)
	veracDPS := float64(maxHit+1) / 2.0 / (float64(atkSpeed) * SecondsPerTick) // 100% accuracy, +1 max
	return 0.75*normalDPS + 0.25*veracDPS
}

// ── Keris partisan ─────────────────────────────────────────────────

// KerisExpectedMax computes expected max hit with Keris proc.
// 50/51 normal, 1/51 triple damage.
func KerisExpectedMax(baseMax int) float64 {
	normal := float64(baseMax)
	proc := float64(baseMax) * 3.0
	return (50.0*normal + 1.0*proc) / 51.0
}

// ── Bolt effects ───────────────────────────────────────────────────

// BoltEffect represents a crossbow bolt enchantment effect.
type BoltEffect struct {
	Name       string
	ProcChance float64 // base chance before Kandarin diary
	// Apply returns the modified (min, max) hit for this bolt proc
	Apply func(baseMax int, monster *data.Monster, rangedLevel int, zcb bool) (int, int)
}

// KandarinDiaryFactor multiplies bolt proc chance by 1.1 if diary is done.
const KandarinDiaryFactor = 1.1

var BoltEffects = map[string]BoltEffect{
	"opal": {
		Name:       "Opal bolts (e)",
		ProcChance: 0.05,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			divisor := 10
			if zcb {
				divisor = 9
			}
			bonus := rngLvl / divisor
			return 0, baseMax + bonus
		},
	},
	"pearl": {
		Name:       "Pearl bolts (e)",
		ProcChance: 0.06,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			divisor := 20
			// Fiery monsters: divisor is 15
			if monsterIsFirey(m) {
				divisor = 15
			}
			if zcb {
				divisor -= 2
			}
			bonus := rngLvl / divisor
			return 0, baseMax + bonus
		},
	},
	"diamond": {
		Name:       "Diamond bolts (e)",
		ProcChance: 0.10,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			mult := 115
			if zcb {
				mult = 126
			}
			return 0, baseMax * mult / 100
		},
	},
	"dragonstone": {
		Name:       "Dragonstone bolts (e)",
		ProcChance: 0.06,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			// Immune if dragon or fiery
			if monsterHasAttr(m, "dragon") || monsterIsFirey(m) {
				return 0, baseMax // no effect
			}
			divisor := 10
			if zcb {
				divisor = 9
			}
			bonus := rngLvl * 2 / divisor
			return 0, baseMax + bonus
		},
	},
	"onyx": {
		Name:       "Onyx bolts (e)",
		ProcChance: 0.11,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			// Immune if undead
			if monsterHasAttr(m, "undead") {
				return 0, baseMax
			}
			mult := 120
			if zcb {
				mult = 132
			}
			return 0, baseMax * mult / 100
		},
	},
	"ruby": {
		Name:       "Ruby bolts (e)",
		ProcChance: 0.06,
		Apply: func(baseMax int, m *data.Monster, rngLvl int, zcb bool) (int, int) {
			pct := 20
			cap := 100
			if zcb {
				pct = 22
				cap = 110
			}
			damage := m.Skills.HP * pct / 100
			if damage > cap {
				damage = cap
			}
			return damage, damage // fixed damage, not a range
		},
	},
}

// BoltExpectedDPS computes DPS with bolt effects factored in.
func BoltExpectedDPS(accuracy float64, minHit, maxHit int, atkSpeed int,
	boltName string, monster *data.Monster, rangedLevel int, zcb bool, kandarinDiary bool) float64 {

	normalDPS := accuracy * float64(maxHit+minHit) / 2.0 / (float64(atkSpeed) * SecondsPerTick)

	bolt, ok := BoltEffects[boltName]
	if !ok {
		return normalDPS
	}

	procChance := bolt.ProcChance
	if kandarinDiary {
		procChance *= KandarinDiaryFactor
	}

	procMin, procMax := bolt.Apply(maxHit, monster, rangedLevel, zcb)
	procDPS := accuracy * float64(procMax+procMin) / 2.0 / (float64(atkSpeed) * SecondsPerTick)

	return (1.0-procChance)*normalDPS + procChance*procDPS
}

// ── NPC Transforms (post-roll damage modifiers) ────────────────────

// ApplyNPCTransform applies monster-specific damage transforms.
// Returns the modified expected damage per hit.
func ApplyNPCTransform(damage float64, monster *data.Monster, style CombatStyle, weapon *data.Equipment) float64 {
	name := monster.Name

	// Zulrah: cap at 50, reroll 45-50
	if strings.Contains(name, "Zulrah") {
		if damage > 50 {
			damage = 47.5 // average of 45-50
		}
	}

	// Verzik P1 (no Dawnbringer): melee capped at 10, else capped at 3
	if strings.Contains(name, "Verzik") && (monster.Version == "Phase 1" || monster.Version == "Phase 1 (story)") {
		if weapon != nil && weapon.Name != "Dawnbringer" {
			if style == StyleStab || style == StyleSlash || style == StyleCrush {
				if damage > 10 {
					damage = 5 // average of linearMin(10)
				}
			} else if damage > 3 {
				damage = 1.5 // average of linearMin(3)
			}
		}
	}

	// Tekton + magic: divide by 5
	if strings.Contains(name, "Tekton") && style == StyleMagic {
		damage /= 5
	}

	// Corporeal Beast + non-corpbane: halve damage
	if strings.Contains(name, "Corporeal Beast") {
		if weapon != nil && !isCorpbaneWeapon(weapon, style) {
			damage /= 2
		}
	}

	// Olm melee/head + magic: divide by 3
	if strings.Contains(name, "Great Olm") {
		if style == StyleMagic && (strings.Contains(monster.Version, "Melee hand") || strings.Contains(monster.Version, "Head")) {
			damage /= 3
		}
		if style == StyleRanged && (strings.Contains(monster.Version, "Mage hand") || strings.Contains(monster.Version, "Melee hand")) {
			damage /= 3
		}
	}

	// Kraken + ranged: /7
	if (strings.Contains(name, "Kraken") || strings.Contains(name, "Cave kraken")) && style == StyleRanged {
		damage /= 7
		if damage < 1 && damage > 0 {
			damage = 1
		}
	}

	// Ice Demon + non-fire non-demonbane: /3
	if ContainsID(IceDemonIDs, monster.ID) && style != StyleMagic {
		damage /= 3
	}

	// Nightmare totems + magic: *2
	if ContainsID(NightmareTotemIDs, monster.ID) && style == StyleMagic {
		damage *= 2
	}

	// Abyssal Sire transition: /2
	if ContainsID(AbyssalSireTransitionIDs, monster.ID) {
		damage /= 2
	}

	// Tormented Demon shielded: *4/5
	if ContainsID(TormentedDemonIDs, monster.ID) &&
		(monster.Version == "Shielded" || monster.Version == "Shielded (Defenceless)") {
		damage = damage * 4 / 5
		if damage < 1 {
			damage = 1
		}
	}

	// Slagilith + non-pickaxe melee: /3
	if strings.Contains(name, "Slagilith") && style != StyleMagic {
		if weapon != nil && !strings.Contains(weapon.Name, "pickaxe") {
			damage /= 3
		}
	}

	// Hueycoatl tail
	if ContainsID(HueycoatlTailIDs, monster.ID) {
		if style == StyleCrush {
			damage = damage * 9 / 10 // linearMin approximation
		} else {
			damage = damage * 4 / 10 // linearMin(4) approximation
		}
	}

	// Hueycoatl head/tail with Pillar phase: *13/10
	if ContainsID(HueycoatlPhaseIDs, monster.ID) && monster.Version == "With Pillar" {
		damage = damage * 13 / 10
	}

	// Flat armour (non-magic)
	if monster.Defensive.FlatArmour > 0 && style != StyleMagic {
		damage -= float64(monster.Defensive.FlatArmour)
		if damage < 0 {
			damage = 0
		}
	}

	return damage
}

// ── Special attack modifiers ───────────────────────────────────────

// SpecialAttackMod returns (accuracy multiplier, max hit multiplier) for a weapon's spec.
type SpecMod struct {
	AccFactor  Factor // multiplied to attack roll
	DmgFactor  Factor // multiplied to max hit
	MultiHits  int    // number of hits (0 = standard single hit)
}

// GetSpecialAttackMod returns the spec modifiers for a weapon.
// Returns nil if the weapon has no special attack or it's not implemented.
func GetSpecialAttackMod(weaponName string) *SpecMod {
	switch {
	// Godswords
	case strings.Contains(weaponName, "Armadyl godsword"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{5, 4}}
	case strings.Contains(weaponName, "Bandos godsword"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{121, 100}} // 1.1 * 1.1
	case strings.Contains(weaponName, "Saradomin godsword"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{121, 100}}
	case strings.Contains(weaponName, "Zamorak godsword"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{11, 10}}

	// Dragon weapons
	case weaponName == "Dragon dagger":
		return &SpecMod{AccFactor: Factor{23, 20}, DmgFactor: Factor{23, 20}, MultiHits: 2}
	case weaponName == "Dragon warhammer":
		return &SpecMod{AccFactor: Factor{1, 1}, DmgFactor: Factor{3, 2}}
	case weaponName == "Dragon claws":
		return &SpecMod{AccFactor: Factor{1, 1}, DmgFactor: Factor{1, 1}, MultiHits: 4}
	case weaponName == "Dragon halberd":
		return &SpecMod{AccFactor: Factor{1, 1}, DmgFactor: Factor{11, 10}, MultiHits: 2}
	case weaponName == "Dragon mace":
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{3, 2}}
	case weaponName == "Dragon sword":
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{5, 4}}
	case weaponName == "Dragon longsword":
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{5, 4}}

	// Other melee
	case weaponName == "Abyssal dagger":
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{17, 20}, MultiHits: 2}
	case weaponName == "Granite hammer":
		return &SpecMod{AccFactor: Factor{3, 2}, DmgFactor: Factor{1, 1}} // +5 flat handled separately
	case weaponName == "Barrelchest anchor":
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{110, 100}}
	case weaponName == "Elder maul":
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{1, 1}}

	// Ranged
	case strings.Contains(weaponName, "Toxic blowpipe"), strings.Contains(weaponName, "Blazing blowpipe"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{3, 2}}
	case strings.Contains(weaponName, "Zaryte crossbow"):
		return &SpecMod{AccFactor: Factor{2, 1}, DmgFactor: Factor{1, 1}}
	case strings.Contains(weaponName, "Heavy ballista"), strings.Contains(weaponName, "Light ballista"):
		return &SpecMod{AccFactor: Factor{5, 4}, DmgFactor: Factor{5, 4}}
	case strings.Contains(weaponName, "Dark bow"):
		return &SpecMod{AccFactor: Factor{1, 1}, DmgFactor: Factor{3, 2}, MultiHits: 2}

	// Magic
	case strings.Contains(weaponName, "Volatile nightmare staff"):
		return &SpecMod{AccFactor: Factor{3, 2}, DmgFactor: Factor{1, 1}}
	case strings.Contains(weaponName, "Accursed sceptre"):
		return &SpecMod{AccFactor: Factor{3, 2}, DmgFactor: Factor{3, 2}}

	default:
		return nil
	}
}

// ── Immunity checks ────────────────────────────────────────────────

// IsImmune checks if the monster is immune to the given combat style.
func IsImmune(monster *data.Monster, style CombatStyle, weapon *data.Equipment) bool {
	// Magic immune NPCs
	if style == StyleMagic && isImmuneToMagic(monster) {
		return true
	}

	// Ranged immune NPCs
	if style == StyleRanged && isImmuneToRanged(monster) {
		return true
	}

	// Melee immune NPCs
	if (style == StyleStab || style == StyleSlash || style == StyleCrush) && isImmuneToMelee(monster) {
		return true
	}

	// Flying + melee
	if monsterHasAttr(monster, "aviansie") && style != StyleMagic && style != StyleRanged {
		return true
	}

	return false
}

// ── Helper functions ───────────────────────────────────────────────

func monsterHasAttr(m *data.Monster, attr string) bool {
	lower := strings.ToLower(attr)
	for _, a := range m.Attributes {
		if strings.ToLower(a) == lower {
			return true
		}
	}
	return false
}

func monsterIsFirey(m *data.Monster) bool {
	return monsterHasAttr(m, "fiery")
}

func isCorpbaneWeapon(weapon *data.Equipment, style CombatStyle) bool {
	name := weapon.Name
	if style == StyleMagic {
		return true
	}
	if style == StyleStab {
		if strings.Contains(name, "spear") || strings.Contains(name, "halberd") ||
			strings.Contains(name, "fang") {
			return true
		}
	}
	return false
}

func isImmuneToMagic(m *data.Monster) bool {
	return ContainsID(ImmuneToMagicDamageIDs, m.ID)
}

func isImmuneToRanged(m *data.Monster) bool {
	return ContainsID(ImmuneToRangedDamageIDs, m.ID)
}

func isImmuneToMelee(m *data.Monster) bool {
	return ContainsID(ImmuneToMeleeDamageIDs, m.ID) ||
		ContainsID(ImmuneToNonSalamanderMeleeIDs, m.ID)
}
