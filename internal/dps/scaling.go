package dps

// Raid monster scaling.
// Ported from weirdgloop/osrs-dps-calc scaling/*.ts

import (
	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// ── Tombs of Amascut ───────────────────────────────────────────────

// ToA monster IDs (from constants.ts TOMBS_OF_AMASCUT_MONSTER_IDS)
var toaMonsterIDs = []int{
	11706, 11707, 11708, 11709, 11710, 11711, 11712, // Baboons
	11713, 11714, 11715, 11716, 11717, 11718, 11719, // More baboons
	11730, 11731, 11732, 11733, // Kephri
	11749, 11750, 11751, 11752, 11753, 11754, 11755, // Akkha
	11756, 11757, 11758, 11759, 11760, 11761, 11762, // Ba-Ba
	11763, 11764, 11765, 11766, 11767, 11768, 11769, // Zebak
	11770, 11771, 11772, 11773, 11774, 11775, 11776, // Wardens P1
	11777, 11778, 11779, 11780, 11781, 11782, 11783, // Wardens P2
	11784, 11785, 11786, 11787, 11788, 11789, 11790, // Wardens P3
	11791, 11792, 11793, 11794, 11795, 11796, 11797, // More wardens
}

// Kephri overlord IDs (excluded from ToA defence scaling)
var kephriOverlordIDs = []int{11731, 11732}

// IsToAMonster checks if a monster is from Tombs of Amascut.
func IsToAMonster(id int) bool {
	for _, mid := range toaMonsterIDs {
		if id == mid {
			return true
		}
	}
	return false
}

// IsKephriOverlord checks if a monster is a Kephri overlord.
func IsKephriOverlord(id int) bool {
	for _, mid := range kephriOverlordIDs {
		if id == mid {
			return true
		}
	}
	return false
}

// ScaleToADefenceRoll scales a ToA monster's defence roll based on invocation level.
// Formula: trunc(defRoll * (250 + invocationLevel) / 250)
func ScaleToADefenceRoll(defRoll, invocationLevel int) int {
	if invocationLevel <= 0 {
		return defRoll
	}
	return defRoll * (250 + invocationLevel) / 250
}

// ── Chambers of Xeric ──────────────────────────────────────────────

// ScaleCoXHP scales a CoX monster's HP based on party size and CM flag.
// Formula from ChambersOfXeric.ts:
//   baseHp * (1 + floor(partySize * 7 / 10 - 1)) for normal
//   baseHp * (1 + floor(partySize * 7 / 10 - 1)) * 3/2 for CM
func ScaleCoXHP(baseHP, partySize int, isCM bool) int {
	if partySize <= 1 {
		if isCM {
			return baseHP * 3 / 2
		}
		return baseHP
	}

	scaled := baseHP * (10 + (partySize*7 - 10)) / 10
	if isCM {
		scaled = scaled * 3 / 2
	}
	return scaled
}

// ── Theatre of Blood ───────────────────────────────────────────────

// ScaleToBHP scales a ToB monster's HP based on party size.
// Formula from TheatreOfBlood.ts:
//   partySize 1-3: HP * 75/100
//   partySize 4: HP * 875/1000
//   partySize 5: HP (full)
func ScaleToBHP(baseHP, partySize int) int {
	switch {
	case partySize <= 3:
		return baseHP * 75 / 100
	case partySize == 4:
		return baseHP * 875 / 1000
	default:
		return baseHP
	}
}

// ── Vardorvis Defence Scaling ──────────────────────────────────────

// Vardorvis defence scales based on current HP percentage.
// At 100% HP: normal defence. At lower HP: reduced defence.
// Formula: defRoll * (currentHP / maxHP)
func ScaleVardorvisDefence(defRoll, currentHP, maxHP int) int {
	if maxHP <= 0 || currentHP >= maxHP {
		return defRoll
	}
	return defRoll * currentHP / maxHP
}

// ── Monster Scaling Entry Point ────────────────────────────────────

// ScaleMonsterForCalc applies any raid-specific scaling to the monster.
// This modifies the monster's effective stats for the calculation.
func ScaleMonsterForCalc(m *data.Monster, opts ScaleOpts) data.Monster {
	scaled := *m // copy

	// ToA HP scaling is applied at data load time by the calc
	// We handle defence scaling here
	if IsToAMonster(m.ID) && !IsKephriOverlord(m.ID) && opts.ToAInvocationLevel > 0 {
		// Defence scaling is applied to the defence roll, not the monster stats
		// So we just store the invocation level for later use
	}

	// CoX HP scaling
	if opts.CoXPartySize > 0 {
		scaled.Skills.HP = ScaleCoXHP(m.Skills.HP, opts.CoXPartySize, opts.CoXCM)
	}

	// ToB HP scaling
	if opts.ToBPartySize > 0 {
		scaled.Skills.HP = ScaleToBHP(m.Skills.HP, opts.ToBPartySize)
	}

	return scaled
}

// ScaleOpts holds raid-specific scaling options.
type ScaleOpts struct {
	ToAInvocationLevel int
	CoXPartySize       int
	CoXCM              bool
	ToBPartySize       int
}
