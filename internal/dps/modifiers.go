package dps

// Additional weapon/monster modifiers.
// Ported from weirdgloop/osrs-dps-calc PlayerVsNPCCalc.ts

import (
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// ── Demonbane Factor ───────────────────────────────────────────────

// DemonbaneFactor computes the additive damage/accuracy bonus against demons.
// vulnerability: 100 default, 70 for Duke Sucellus, 120 for Yama, 200 for Yama Void Flare.
// weaponPercent: base percentage for the weapon (e.g., 70 for Arclight).
// Returns [percent, 100] factor for use with ApplyAddFactor.
func DemonbaneFactor(weaponPercent, vulnerability int) Factor {
	if vulnerability <= 0 {
		vulnerability = 100
	}
	percent := weaponPercent * vulnerability / 100
	return Factor{percent, 100}
}

// GetDemonbaneVulnerability returns the demonbane vulnerability for a monster.
func GetDemonbaneVulnerability(m *data.Monster) int {
	switch {
	case m.Name == "Duke Sucellus":
		return 70
	case isYama(m):
		return 120
	case isYamaVoidFlare(m):
		return 200
	default:
		return 100
	}
}

// Yama IDs
func isYama(m *data.Monster) bool {
	yamaIDs := []int{13672, 13673, 13674}
	for _, id := range yamaIDs {
		if m.ID == id {
			return true
		}
	}
	return false
}

func isYamaVoidFlare(m *data.Monster) bool {
	flareIDs := []int{13675, 13676, 13677}
	for _, id := range flareIDs {
		if m.ID == id {
			return true
		}
	}
	return false
}

// ── Spell Element Weakness ─────────────────────────────────────────

// SpellElementWeaknessBonus computes the bonus attack roll or damage
// from matching a spell's element to the monster's weakness.
// Returns the additive bonus (trunc(baseRoll * severity / 100)).
func SpellElementWeaknessBonus(baseValue int, monster *data.Monster, spellElement string) int {
	if monster.Weakness == nil || spellElement == "" {
		return 0
	}
	if strings.ToLower(monster.Weakness.Element) != strings.ToLower(spellElement) {
		return 0
	}
	return baseValue * monster.Weakness.Severity / 100
}

// ── Salve Amulet / Slayer Helm stacking ────────────────────────────

// These are mutually exclusive in the TypeScript source.
// Salve takes priority over slayer helm.
// For ranged with imbued black mask, the numerator can stack additively
// with rev weapon (+10), dragonbane (+5), and scorching bow (+6).

// ── Tormented Demon ────────────────────────────────────────────────

// TormentedDemonUnshieldedBonus: +max(0, atkSpeed^2 - 16) flat damage on accurate hits.
func TormentedDemonUnshieldedBonus(atkSpeed int) int {
	bonus := atkSpeed*atkSpeed - 16
	if bonus < 0 {
		return 0
	}
	return bonus
}

// TormentedDemonShieldedMod: damage * 4/5 with floor 1.
func TormentedDemonShieldedMod(damage int) int {
	result := damage * 4 / 5
	if result < 1 && damage > 0 {
		return 1
	}
	return result
}

// ── Vampire Damage Modifiers ───────────────────────────────────────

// VampyreDamageModifier returns the damage multiplier for vampyre monsters.
// Depends on weapon type and Efaritay's aid.
func VampyreDamageModifier(weaponName string, hasEfaritay bool) Factor {
	lower := strings.ToLower(weaponName)

	if !hasEfaritay {
		return Factor{1, 1} // no bonus without Efaritay's
	}

	switch {
	case strings.Contains(lower, "blisterwood flail"):
		// * 11/10 then * 5/4 = 55/40 = 11/8
		return Factor{55, 40}
	case strings.Contains(lower, "blisterwood sickle"):
		// * 11/10 then * 23/20 = 253/200
		return Factor{253, 200}
	case strings.Contains(lower, "ivandis flail"):
		// * 11/10 then * 6/5 = 66/50 = 33/25
		return Factor{33, 25}
	default:
		// Silver weapon + Efaritay's: * 11/10 then * 11/10 = 121/100
		return Factor{121, 100}
	}
}

// ── Berserker Necklace ─────────────────────────────────────────────

// BerserkerNecklaceFactor: * 6/5 when wearing berserker necklace + tzhaar weapon.
func BerserkerNecklaceFactor() Factor {
	return Factor{6, 5}
}

// ── Leaf-bladed Battleaxe ──────────────────────────────────────────

// LeafBladedBattleaxeFactor: * 47/40 vs leafy monsters.
func LeafBladedBattleaxeFactor() Factor {
	return Factor{47, 40}
}

// ── Barronite Mace ─────────────────────────────────────────────────

// BarroniteMaceFactor: * 23/20 vs golems.
func BarroniteMaceFactor() Factor {
	return Factor{23, 20}
}

// ── Keris Partisan Variants ────────────────────────────────────────

// KerisBreachingFactor: * 133/100 accuracy and damage vs kalphites.
func KerisBreachingFactor() Factor {
	return Factor{133, 100}
}

// KerisAmascutFactor: * 115/100 damage vs kalphites (weaker variant).
func KerisAmascutFactor() Factor {
	return Factor{115, 100}
}

// KerisSunFactor: * 5/4 accuracy when target HP < 25% in ToA.
func KerisSunFactor() Factor {
	return Factor{5, 4}
}

// ── Smoke Battlestaff ──────────────────────────────────────────────

// SmokeBattlestaffAccuracyBonus: +10% to magic attack roll (additive).
const SmokeBattlestaffAccuracyBonus = 10

// SmokeBattlestaffDamageBonus: +100 to magic damage bonus (10.0%).
const SmokeBattlestaffDamageBonus = 100

// ── Tome Bonuses ───────────────────────────────────────────────────

// TomeDamageFactor: * 11/10 for matching tome + spell element.
func TomeDamageFactor() Factor {
	return Factor{11, 10}
}

// ── Guardian (CoX) Pickaxe Scaling ─────────────────────────────────

// GuardianPickaxeBonus returns the mining bonus for guardians in CoX.
func GuardianPickaxeBonus(pickaxeTier string) int {
	switch strings.ToLower(pickaxeTier) {
	case "bronze", "iron":
		return 1
	case "steel":
		return 6
	case "black":
		return 11
	case "mithril":
		return 21
	case "adamant":
		return 31
	case "rune", "gilded":
		return 41
	case "dragon", "dragon (or)", "infernal", "crystal", "3rd age":
		return 61
	default:
		return 1
	}
}

// GuardianDamageTransform: damage = trunc((50 + miningLevel + pickBonus) * damage / 150).
func GuardianDamageTransform(damage, miningLevel int, pickaxeTier string) int {
	pickBonus := GuardianPickaxeBonus(pickaxeTier)
	return (50 + miningLevel + pickBonus) * damage / 150
}
