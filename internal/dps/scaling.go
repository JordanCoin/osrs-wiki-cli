package dps

// Raid monster scaling.
// Ported from weirdgloop/osrs-dps-calc scaling/*.ts

import (
	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// ── Tombs of Amascut ───────────────────────────────────────────────

// IsToAMonster checks if a monster is from Tombs of Amascut.
// Uses the canonical ID list from constants.go.
func IsToAMonster(id int) bool {
	return ContainsID(ToAMonsterIDs, id)
}

// IsKephriOverlord checks if a monster is a Kephri overlord.
// Uses the canonical ID list from constants.go.
func IsKephriOverlord(id int) bool {
	return ContainsID(KephriOverlordIDs, id)
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

// ── Vardorvis Scaling ──────────────────────────────────────────────

// VardorvisScaling holds the HP-dependent stat ranges for Vardorvis.
type VardorvisScaling struct {
	MaxHP    int
	StrRange [2]int // [start at full HP, end at 0 HP]
	DefRange [2]int // [start at full HP, end at 0 HP]
}

// GetVardorvisScaling returns the scaling parameters for a Vardorvis version.
func GetVardorvisScaling(version string) VardorvisScaling {
	switch version {
	case "Quest":
		return VardorvisScaling{MaxHP: 500, StrRange: [2]int{210, 280}, DefRange: [2]int{180, 130}}
	case "Awakened":
		return VardorvisScaling{MaxHP: 1400, StrRange: [2]int{391, 522}, DefRange: [2]int{268, 181}}
	default: // Post-quest
		return VardorvisScaling{MaxHP: 700, StrRange: [2]int{270, 360}, DefRange: [2]int{215, 145}}
	}
}

// ApplyVardorvisScaling applies HP-dependent stat scaling to Vardorvis.
// At full HP: def = DefRange[0] (e.g., 215 for post-quest).
// At 0 HP: def = DefRange[1] (e.g., 145 for post-quest).
// For DPS calculations, we use full HP (worst case defence).
func ApplyVardorvisScaling(m *data.Monster, currentHP int) data.Monster {
	if m.Name != "Vardorvis" {
		return *m
	}

	scaling := GetVardorvisScaling(m.Version)
	if currentHP < 0 {
		currentHP = scaling.MaxHP // default to full HP
	}

	scaled := *m
	scaled.Skills.Str = lerp(currentHP, scaling.MaxHP, 0, scaling.StrRange[0], scaling.StrRange[1])
	scaled.Skills.Def = lerp(currentHP, scaling.MaxHP, 0, scaling.DefRange[0], scaling.DefRange[1])
	return scaled
}

// lerp: linear interpolation with integer truncation.
// lerp(curr, srcStart, srcEnd, dstStart, dstEnd)
func lerp(curr, srcStart, srcEnd, dstStart, dstEnd int) int {
	srcRange := srcEnd - srcStart
	if srcRange == 0 {
		return dstStart
	}
	dstRange := dstEnd - dstStart
	currNorm := curr - srcStart
	return currNorm*dstRange/srcRange + dstStart
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
