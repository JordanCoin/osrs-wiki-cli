package dps

// 1-to-1 port of weirdgloop/osrs-dps-calc PlayerVsNPCCalc.ts
// Source: https://github.com/weirdgloop/osrs-dps-calc
// All arithmetic uses integer truncation (Go's default int division).

import (
	"fmt"
	"math"
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// CombatStyle represents a player's attack style.
type CombatStyle string

const (
	StyleStab   CombatStyle = "stab"
	StyleSlash  CombatStyle = "slash"
	StyleCrush  CombatStyle = "crush"
	StyleRanged CombatStyle = "ranged"
	StyleMagic  CombatStyle = "magic"
)

const SecondsPerTick = 0.6

// Factor represents an integer fraction [numerator, denominator].
// Used for prayer bonuses and multipliers to match TypeScript's integer math.
type Factor [2]int

// ApplyFactor computes trunc(base * factor[0] / factor[1]).
func ApplyFactor(base int, f Factor) int {
	return base * f[0] / f[1]
}

// ApplyAddFactor computes base + trunc(base * factor[0] / factor[1]).
func ApplyAddFactor(base int, f Factor) int {
	addend := base * f[0] / f[1]
	return base + addend
}

// MaxHitFromEffective computes trunc((effectiveLevel * gearBonus + 320) / 640).
func MaxHitFromEffective(effectiveLevel, gearBonus int) int {
	return (effectiveLevel*gearBonus + 320) / 640
}

// Prayer data matching TypeScript PrayerMap.
// Factors are [numerator, denominator] where denominator is always 100.
type PrayerData struct {
	Name           string
	CombatStyle    string // "melee", "ranged", "magic"
	FactorAccuracy Factor // [0,0] means no bonus
	FactorStrength Factor
	MagicDmgBonus  int // added to magic damage % (in tenths)
}

var Prayers = map[string]PrayerData{
	"piety":         {Name: "Piety", CombatStyle: "melee", FactorAccuracy: Factor{120, 100}, FactorStrength: Factor{123, 100}},
	"chivalry":      {Name: "Chivalry", CombatStyle: "melee", FactorAccuracy: Factor{115, 100}, FactorStrength: Factor{118, 100}},
	"rigour":        {Name: "Rigour", CombatStyle: "ranged", FactorAccuracy: Factor{120, 100}, FactorStrength: Factor{123, 100}},
	"deadeye":       {Name: "Deadeye", CombatStyle: "ranged", FactorAccuracy: Factor{118, 100}, FactorStrength: Factor{118, 100}},
	"augury":        {Name: "Augury", CombatStyle: "magic", FactorAccuracy: Factor{125, 100}, FactorStrength: Factor{0, 1}, MagicDmgBonus: 40},
	"mystic_vigour": {Name: "Mystic Vigour", CombatStyle: "magic", FactorAccuracy: Factor{118, 100}, FactorStrength: Factor{0, 1}, MagicDmgBonus: 30},
	"mystic_might":  {Name: "Mystic Might", CombatStyle: "magic", FactorAccuracy: Factor{115, 100}, FactorStrength: Factor{0, 1}, MagicDmgBonus: 20},
	"eagle_eye":     {Name: "Eagle Eye", CombatStyle: "ranged", FactorAccuracy: Factor{115, 100}, FactorStrength: Factor{115, 100}},
	"none":          {Name: "None", CombatStyle: "", FactorAccuracy: Factor{1, 1}, FactorStrength: Factor{1, 1}},
}

// PlayerStats represents the player's combat levels + boosts.
type PlayerStats struct {
	Attack    int
	Strength  int
	Defence   int
	Ranged    int
	Magic     int
	Prayer    int
	Hitpoints int
	// Boosts from potions
	AtkBoost int
	StrBoost int
	RngBoost int
	MagBoost int
}

// MaxedStats returns maxed combat stats with super combat + ranging potion boosts.
func MaxedStats() PlayerStats {
	return PlayerStats{
		Attack: 99, Strength: 99, Defence: 99,
		Ranged: 99, Magic: 99, Prayer: 99, Hitpoints: 99,
		AtkBoost: 19, StrBoost: 19, RngBoost: 13, MagBoost: 0,
	}
}

// MaxedStatsNoPot returns maxed stats with no potion boosts.
func MaxedStatsNoPot() PlayerStats {
	return PlayerStats{
		Attack: 99, Strength: 99, Defence: 99,
		Ranged: 99, Magic: 99, Prayer: 99, Hitpoints: 99,
	}
}

// TotalEquipmentBonuses holds summed equipment stats.
type TotalEquipmentBonuses struct {
	StabAtk   int
	SlashAtk  int
	CrushAtk  int
	MagicAtk  int
	RangedAtk int
	MeleeStr  int
	RangedStr int
	MagicStr  int // in tenths of a percent (e.g., 150 = 15.0%)
}

// DPSResult holds the output of a DPS calculation.
type DPSResult struct {
	Weapon    string      `json:"weapon"`
	Style     CombatStyle `json:"style"`
	Prayer    string      `json:"prayer"`
	MaxHit    int         `json:"max_hit"`
	MinHit    int         `json:"min_hit"`
	Accuracy  float64     `json:"accuracy"`
	DPS       float64     `json:"dps"`
	TTKTicks  int         `json:"ttk_ticks"`
	TTKString string      `json:"ttk_string"`
	MonsterHP int         `json:"monster_hp"`
	AtkSpeed  int         `json:"attack_speed"`
	AtkRoll   int         `json:"atk_roll"`
	DefRoll   int         `json:"def_roll"`
}

// CalcContext holds all state for a single DPS calculation.
type CalcContext struct {
	Stats   PlayerStats
	Bonuses TotalEquipmentBonuses
	Weapon  *data.Equipment
	Gear    GearSet
	Monster *data.Monster
	Prayer  PrayerData
	Style   CombatStyle
	// All equipped item names (lowercase) for quick checks
	AllItems []string
}

// Calculate computes DPS for a gear set against a monster.
// Follows the same order of operations as TypeScript PlayerVsNPCCalc.
func Calculate(gear GearSet, monster *data.Monster, stats PlayerStats) (*DPSResult, error) {
	bonuses, weapon, err := resolveGearBonuses(gear)
	if err != nil {
		return nil, err
	}
	if weapon == nil {
		return nil, fmt.Errorf("no weapon found in gear set")
	}

	style, prayer := determineCombatParams(gear, weapon)

	// Apply Vardorvis HP-dependent stat scaling
	scaledMonster := ApplyVardorvisScaling(monster, -1) // -1 = full HP

	ctx := &CalcContext{
		Stats:   stats,
		Bonuses: bonuses,
		Weapon:  weapon,
		Gear:    gear,
		Monster: &scaledMonster,
		Prayer:  prayer,
		Style:   style,
	}
	// Build lowercase item list for wearing() checks
	for slot, name := range gear.GearSlots() {
		_ = slot
		if name != "" {
			ctx.AllItems = append(ctx.AllItems, strings.ToLower(name))
		}
	}

	// ── Step 0: Immunity check ─────────────────────────────────────
	if IsImmune(monster, style, weapon) {
		return &DPSResult{
			Weapon: weapon.Name, Style: style, Prayer: prayer.Name,
			TTKString: "immune", MonsterHP: monster.Skills.HP,
			AtkSpeed: max(weapon.Speed, 4),
		}, nil
	}

	// ── Step 1: One-hit monsters ───────────────────────────────────
	if ContainsID(OneHitMonsters, monster.ID) {
		return &DPSResult{
			Weapon: weapon.Name, Style: style, Prayer: prayer.Name,
			MaxHit: monster.Skills.HP, Accuracy: 100.0,
			DPS: float64(monster.Skills.HP) / (float64(max(weapon.Speed, 4)) * SecondsPerTick),
			TTKTicks: max(weapon.Speed, 4), TTKString: formatTTK(max(weapon.Speed, 4)),
			MonsterHP: monster.Skills.HP, AtkSpeed: max(weapon.Speed, 4),
		}, nil
	}

	// ── Step 2: Attack and defence rolls ───────────────────────────
	atkRoll := ctx.getPlayerMaxAttackRoll()
	defRoll := ctx.getNPCDefenceRoll()

	// ToA defence scaling
	if IsToAMonster(monster.ID) && !IsKephriOverlord(monster.ID) {
		// Default invocation level 150 if not specified
		defRoll = ScaleToADefenceRoll(defRoll, 150)
	}

	// ── Step 3: Accuracy ───────────────────────────────────────────
	accuracy := getNormalAccuracyRoll(atkRoll, defRoll)

	// Guaranteed accuracy monsters
	if ContainsID(GuaranteedAccuracyMonsters, monster.ID) {
		accuracy = 1.0
	}

	// P2 Wardens: always 100% accuracy
	if ContainsID(P2WardenIDs, monster.ID) {
		accuracy = 1.0
	}

	// Fang accuracy override (stab style, non-ToA)
	if ctx.isWearingFang() && style == StyleStab {
		if IsToAMonster(monster.ID) {
			// In ToA: 1 - (1-accuracy)^2
			accuracy = 1.0 - (1.0-accuracy)*(1.0-accuracy)
		} else {
			accuracy = getFangAccuracyRoll(atkRoll, defRoll)
		}
	}

	// ── Step 4: Max hit ────────────────────────────────────────────
	minHit, maxHit := ctx.getPlayerMaxHit()

	// Rev weapon buff (wilderness, charged)
	if ctx.isRevWeaponApplicable() {
		revFactor := RevWeaponFactor()
		maxHit = ApplyFactor(maxHit, revFactor)
		atkRoll = ApplyFactor(atkRoll, revFactor)
		// Recalculate accuracy with boosted attack roll
		accuracy = getNormalAccuracyRoll(atkRoll, defRoll)
	}

	// Keris + kalphite
	if ctx.wearingAny("keris") && ctx.monsterHasAttribute("kalphite") {
		if ctx.wearing("Keris partisan of breaching") {
			maxHit = ApplyFactor(maxHit, KerisBreachingFactor())
			atkRoll = ApplyFactor(atkRoll, KerisBreachingFactor())
			accuracy = getNormalAccuracyRoll(atkRoll, defRoll)
		}
	}

	// ── Step 5: Attack speed ───────────────────────────────────────
	atkSpeed := weapon.Speed
	if atkSpeed <= 0 {
		atkSpeed = 4
	}

	// ── Step 6: Special weapon distributions ───────────────────────
	var expectedDmg float64
	specialHandled := false

	// Scythe of vitur: multi-hitsplat
	if ctx.wearingAny("of vitur") || ctx.wearing("Scythe of vitur") {
		scytheMax := ScytheExpectedMax(maxHit, monster.Size)
		expectedDmg = accuracy * float64(scytheMax) / 2.0
		specialHandled = true
	}

	// Dharok's set: scales with missing HP
	if !specialHandled && ctx.isWearingDharok() {
		dharokMax := DharokMaxHit(maxHit, ctx.Stats.Hitpoints, 1) // 1 HP for max DPS
		expectedDmg = accuracy * float64(dharokMax) / 2.0
		specialHandled = true
	}

	// Verac's set: 25% ignores defence
	if !specialHandled && ctx.isWearingVeracs() {
		expectedDmg = VeracExpectedDPS(accuracy, maxHit, atkSpeed) * float64(atkSpeed) * SecondsPerTick
		specialHandled = true
	}

	// Dragon claws spec
	if !specialHandled && ctx.wearing("Dragon claws") && ctx.Gear.UseSpec {
		expectedDmg = DragonClawsExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Burning claws spec
	if !specialHandled && ctx.wearing("Burning claws") && ctx.Gear.UseSpec {
		expectedDmg = BurningClawsExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Dark bow
	if !specialHandled && ctx.wearing("Dark bow") {
		dragonArrows := ctx.wearingAny("dragon arrow")
		expectedDmg = DarkBowExpectedDmg(accuracy, maxHit, dragonArrows)
		specialHandled = true
	}

	// Tonalztics of ralos (charged, ranged)
	if !specialHandled && ctx.wearing("Tonalztics of ralos") && style == StyleRanged {
		expectedDmg = TonalzticsExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Two-hit weapons (Torag's hammers, Sulphur blades, etc.)
	if !specialHandled && ctx.isTwoHitWeapon() {
		expectedDmg = TwoHitExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Dual macuahuitl
	if !specialHandled && ctx.wearing("Dual macuahuitl") {
		expectedDmg = DualMacuahuitlExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Keris + kalphite proc (50/51 normal, 1/51 triple)
	if !specialHandled && ctx.wearingAny("keris") && ctx.monsterHasAttribute("kalphite") {
		kerisExpectedMax := KerisExpectedMax(maxHit)
		expectedDmg = accuracy * kerisExpectedMax / 2.0
		specialHandled = true
	}

	// Karil's set with Amulet of the Damned
	if !specialHandled && ctx.isWearingKarils() {
		expectedDmg = KarilsExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Ahrim's set with Amulet of the Damned
	if !specialHandled && ctx.isWearingAhrims() {
		expectedDmg = AhrimsExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Voidwaker spec: guaranteed accuracy, min=max/2
	if !specialHandled && ctx.wearing("Voidwaker") && ctx.Gear.UseSpec {
		expectedDmg = VoidwakerExpectedDmg(maxHit)
		specialHandled = true
	}

	// P2 Wardens: special min/max calculation
	if !specialHandled && ContainsID(P2WardenIDs, monster.ID) {
		wMinHit, wMaxHit := P2WardensModifier(maxHit, atkRoll, defRoll)
		expectedDmg = float64(wMaxHit+wMinHit) / 2.0 // 100% accuracy
		minHit = wMinHit
		maxHit = wMaxHit
		specialHandled = true
	}

	// Blood Moon set
	if !specialHandled && ctx.isWearingBloodMoon() {
		expectedDmg = BloodMoonExpectedDmg(accuracy, maxHit)
		specialHandled = true
	}

	// Standard single-hit distribution
	if !specialHandled {
		expectedDmg = accuracy * float64(maxHit+minHit) / 2.0
	}

	// ── Step 7: NPC transforms (post-roll) ─────────────────────────
	expectedDmg = ApplyNPCTransform(expectedDmg, monster, style, weapon)

	// Berserker necklace + tzhaar weapon
	if ctx.isWearingBerserkerNecklace() && ctx.isWearingTzhaarWeapon() {
		expectedDmg = expectedDmg * 6.0 / 5.0
	}

	// Vampyre damage with Efaritay's aid
	if ctx.monsterHasAttribute("vampyre") && ctx.wearing("Efaritay's aid") {
		vFactor := VampyreDamageModifier(weapon.Name, true)
		expectedDmg = float64(int(expectedDmg) * vFactor[0] / vFactor[1])
	}

	// ── Step 8: DPS ────────────────────────────────────────────────
	dps := expectedDmg / (float64(atkSpeed) * SecondsPerTick)

	// ── Step 9: TTK (use original monster HP, not scaled) ──────────
	monsterHP := monster.Skills.HP
	if monsterHP == 0 {
		monsterHP = scaledMonster.Skills.HP
	}
	ttkTicks := 0
	if dps > 0 {
		seconds := float64(monsterHP) / dps
		ttkTicks = int(math.Ceil(seconds/SecondsPerTick/float64(atkSpeed))) * atkSpeed
	}

	return &DPSResult{
		Weapon:    weapon.Name,
		Style:     style,
		Prayer:    prayer.Name,
		MaxHit:    maxHit,
		MinHit:    minHit,
		Accuracy:  math.Round(accuracy*1000) / 10,
		DPS:       math.Round(dps*100) / 100,
		TTKTicks:  ttkTicks,
		TTKString: formatTTK(ttkTicks),
		MonsterHP: monster.Skills.HP,
		AtkSpeed:  atkSpeed,
		AtkRoll:   atkRoll,
		DefRoll:   defRoll,
	}, nil
}

// ── Accuracy formulas ──────────────────────────────────────────────

// getNormalAccuracyRoll computes standard accuracy.
// Matches BaseCalc.getNormalAccuracyRoll in TypeScript.
func getNormalAccuracyRoll(atk, def int) float64 {
	a := atk
	d := def

	if a < 0 {
		a = min(0, a+2)
	}
	if d < 0 {
		d = min(0, d+2)
	}

	fa := float64(a)
	fd := float64(d)

	if a >= 0 && d >= 0 {
		if a > d {
			return 1.0 - (fd+2.0)/(2.0*(fa+1.0))
		}
		return fa / (2.0*fd + 1.0)
	}
	if a >= 0 && d < 0 {
		return 1.0 - 1.0/float64(-d+1)/float64(a+1)
	}
	if a < 0 && d >= 0 {
		return 0
	}
	// Both negative
	return getNormalAccuracyRoll(-d, -a)
}

// getFangAccuracyRoll computes Osmumten's fang double-roll accuracy.
func getFangAccuracyRoll(atk, def int) float64 {
	a := atk
	d := def

	if a < 0 {
		a = min(0, a+2)
	}
	if d < 0 {
		d = min(0, d+2)
	}

	fa := float64(a)
	fd := float64(d)

	if a >= 0 && d >= 0 {
		if a > d {
			return 1.0 - (fd+2.0)*(2.0*fd+3.0)/(fa+1.0)/(fa+1.0)/6.0
		}
		return fa * (4.0*fa + 5.0) / 6.0 / (fa + 1.0) / (fd + 1.0)
	}
	if a >= 0 && d < 0 {
		return 1.0 - 1.0/float64(-d+1)/float64(a+1)
	}
	if a < 0 && d >= 0 {
		return 0
	}
	// Both negative: reverse roll
	nd := float64(-a)
	na := float64(-d)
	if na < nd {
		return na * (nd*6.0 - 2.0*na + 5.0) / 6.0 / (nd + 1.0) / (nd + 1.0)
	}
	return 1.0 - (nd+2.0)*(2.0*nd+3.0)/6.0/(nd+1.0)/(na+1.0)
}

// ── NPC Defence Roll ───────────────────────────────────────────────

func (ctx *CalcContext) getNPCDefenceRoll() int {
	m := ctx.Monster
	style := ctx.Style

	// For magic style, most monsters use magic level for defence, not def level.
	// Exception: some specific NPCs use defence level.
	var level int
	if style == StyleMagic && !usesDefLevelForMagicDef(m) {
		level = m.Skills.Magic
	} else {
		level = m.Skills.Def
	}

	effectiveLevel := level + 9

	var bonus int
	switch style {
	case StyleStab:
		bonus = m.Defensive.Stab
	case StyleSlash:
		bonus = m.Defensive.Slash
	case StyleCrush:
		bonus = m.Defensive.Crush
	case StyleMagic:
		bonus = m.Defensive.Magic
	case StyleRanged:
		bonus = m.Defensive.Standard
	}

	defRoll := effectiveLevel * (bonus + 64)
	return defRoll
}

// usesDefLevelForMagicDef checks if a monster uses defence level instead of magic
// level for magic defence calculations.
func usesDefLevelForMagicDef(m *data.Monster) bool {
	return ContainsID(UsesDefLevelForMagicDefIDs, m.ID)
}

// ── Player Max Attack Roll ─────────────────────────────────────────

func (ctx *CalcContext) getPlayerMaxAttackRoll() int {
	switch {
	case ctx.Style == StyleRanged:
		return ctx.getPlayerMaxRangedAttackRoll()
	case ctx.Style == StyleMagic:
		return ctx.getPlayerMaxMagicAttackRoll()
	default:
		return ctx.getPlayerMaxMeleeAttackRoll()
	}
}

func (ctx *CalcContext) getPlayerMaxMeleeAttackRoll() int {
	effectiveLevel := ctx.Stats.Attack + ctx.Stats.AtkBoost

	// Prayer
	if ctx.Prayer.FactorAccuracy[0] != 0 {
		effectiveLevel = ApplyFactor(effectiveLevel, ctx.Prayer.FactorAccuracy)
	}

	// Stance bonus: +8 base, +3 accurate, +1 controlled
	effectiveLevel += 8

	// Void melee: * 11/10
	if ctx.isWearingMeleeVoid() {
		effectiveLevel = effectiveLevel * 11 / 10
	}

	// Base roll
	var equipBonus int
	switch ctx.Style {
	case StyleStab:
		equipBonus = ctx.Bonuses.StabAtk
	case StyleSlash:
		equipBonus = ctx.Bonuses.SlashAtk
	case StyleCrush:
		equipBonus = ctx.Bonuses.CrushAtk
	}
	attackRoll := effectiveLevel * (equipBonus + 64)

	// Salve amulet (e)/(ei) + undead: * 6/5
	if ctx.wearing("Salve amulet(ei)") && ctx.monsterHasAttribute("undead") {
		attackRoll = ApplyFactor(attackRoll, Factor{6, 5})
	} else if ctx.wearing("Salve amulet (e)") && ctx.monsterHasAttribute("undead") {
		attackRoll = ApplyFactor(attackRoll, Factor{6, 5})
	} else if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Slayer helmet") || ctx.wearing("Black mask (i)") || ctx.wearing("Black mask")) && ctx.isOnSlayerTask() {
		// Black mask / slayer helm: * 7/6
		attackRoll = ApplyFactor(attackRoll, Factor{7, 6})
	}

	// Dragon hunter lance + dragon
	if ctx.wearing("Dragon hunter lance") && ctx.monsterHasAttribute("dragon") {
		attackRoll = ApplyFactor(attackRoll, Factor{6, 5})
	}

	// Arclight + demon
	if ctx.wearing("Arclight") && ctx.monsterHasAttribute("demon") {
		attackRoll = ApplyAddFactor(attackRoll, Factor{70, 100})
	}

	// Inquisitor's (crush style)
	if ctx.Style == StyleCrush {
		pieces := ctx.countInquisitorPieces()
		if pieces > 0 {
			if ctx.wearing("Inquisitor's mace") {
				pieces *= 5 // 2.5% each, no full set bonus
			} else if pieces == 3 {
				pieces = 5 // 1% extra for full set
			}
			attackRoll = ApplyFactor(attackRoll, Factor{200 + pieces, 200})
		}
	}

	return attackRoll
}

func (ctx *CalcContext) getPlayerMaxRangedAttackRoll() int {
	effectiveLevel := ctx.Stats.Ranged + ctx.Stats.RngBoost

	if ctx.Prayer.FactorAccuracy[0] != 0 {
		effectiveLevel = ApplyFactor(effectiveLevel, ctx.Prayer.FactorAccuracy)
	}

	effectiveLevel += 8

	// Ranged void: * 11/10
	if ctx.isWearingRangedVoid() {
		effectiveLevel = effectiveLevel * 11 / 10
	}

	attackRoll := effectiveLevel * (ctx.Bonuses.RangedAtk + 64)

	// Crystal bow/Bowfa: * (20 + crystalPieces) / 20
	if ctx.wearing("Crystal bow") || ctx.wearingAny("Bow of faerdhinen") {
		pieces := ctx.countCrystalPieces()
		attackRoll = attackRoll * (20 + pieces) / 20
	}

	// Salve amulet(ei) + undead
	if ctx.wearing("Salve amulet(ei)") && ctx.monsterHasAttribute("undead") {
		attackRoll = ApplyFactor(attackRoll, Factor{6, 5})
	} else if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Black mask (i)")) && ctx.isOnSlayerTask() {
		attackRoll = ApplyFactor(attackRoll, Factor{23, 20})
	}

	// Twisted bow scaling
	if ctx.wearing("Twisted bow") {
		attackRoll = tbowScaling(attackRoll, ctx.Monster, true)
	}

	// Dragon hunter crossbow + dragon
	if ctx.wearing("Dragon hunter crossbow") && ctx.monsterHasAttribute("dragon") {
		attackRoll = ApplyFactor(attackRoll, Factor{13, 10})
	}

	return attackRoll
}

func (ctx *CalcContext) getPlayerMaxMagicAttackRoll() int {
	effectiveLevel := ctx.Stats.Magic + ctx.Stats.MagBoost

	if ctx.Prayer.FactorAccuracy[0] != 0 {
		effectiveLevel = ApplyFactor(effectiveLevel, ctx.Prayer.FactorAccuracy)
	}

	effectiveLevel += 9

	// Magic void: * 29/20
	if ctx.isWearingMagicVoid() {
		effectiveLevel = effectiveLevel * 29 / 20
	}

	attackRoll := effectiveLevel * (ctx.Bonuses.MagicAtk + 64)

	// Salve amulet(ei) + undead: additive +20%
	if ctx.wearing("Salve amulet(ei)") && ctx.monsterHasAttribute("undead") {
		attackRoll = attackRoll * 120 / 100
	} else if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Black mask (i)")) && ctx.isOnSlayerTask() {
		attackRoll = ApplyFactor(attackRoll, Factor{23, 20})
	}

	// Smoke battlestaff + standard spellbook: +10%
	if ctx.wearing("Smoke battlestaff") || ctx.wearing("Mystic smoke staff") {
		attackRoll = attackRoll * 110 / 100
	}

	// Dragon hunter lance/wand + dragon
	if ctx.wearing("Dragon hunter wand") && ctx.monsterHasAttribute("dragon") {
		attackRoll = ApplyFactor(attackRoll, Factor{7, 4})
	}

	return attackRoll
}

// ── Player Max Hit ─────────────────────────────────────────────────

func (ctx *CalcContext) getPlayerMaxHit() (minHit, maxHit int) {
	switch {
	case ctx.Style == StyleRanged:
		return ctx.getPlayerMaxRangedHit()
	case ctx.Style == StyleMagic:
		return ctx.getPlayerMaxMagicHit()
	default:
		return ctx.getPlayerMaxMeleeHit()
	}
}

func (ctx *CalcContext) getPlayerMaxMeleeHit() (int, int) {
	effectiveLevel := ctx.Stats.Strength + ctx.Stats.StrBoost

	// Prayer
	if ctx.Prayer.FactorStrength[0] != 0 {
		effectiveLevel = ApplyFactor(effectiveLevel, ctx.Prayer.FactorStrength)
	}

	// Stance: +8 base, +3 aggressive, +1 controlled
	effectiveLevel += 8

	// Melee void: * 11/10
	if ctx.isWearingMeleeVoid() {
		effectiveLevel = effectiveLevel * 11 / 10
	}

	maxHit := MaxHitFromEffective(effectiveLevel, ctx.Bonuses.MeleeStr+64)

	// Salve amulet(e)/(ei) + undead: * 6/5
	if (ctx.wearing("Salve amulet(ei)") || ctx.wearing("Salve amulet (e)")) && ctx.monsterHasAttribute("undead") {
		maxHit = ApplyFactor(maxHit, Factor{6, 5})
	} else if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Slayer helmet") || ctx.wearing("Black mask (i)") || ctx.wearing("Black mask")) && ctx.isOnSlayerTask() {
		maxHit = ApplyFactor(maxHit, Factor{7, 6})
	}

	// Dragon hunter lance + dragon
	if ctx.wearing("Dragon hunter lance") && ctx.monsterHasAttribute("dragon") {
		maxHit = ApplyFactor(maxHit, Factor{6, 5})
	}

	// Arclight + demon
	if ctx.wearing("Arclight") && ctx.monsterHasAttribute("demon") {
		maxHit = ApplyAddFactor(maxHit, Factor{70, 100})
	}

	// Inquisitor's (crush style)
	if ctx.Style == StyleCrush {
		pieces := ctx.countInquisitorPieces()
		if pieces > 0 {
			if ctx.wearing("Inquisitor's mace") {
				pieces *= 5
			} else if pieces == 3 {
				pieces = 5
			}
			maxHit = ApplyFactor(maxHit, Factor{200 + pieces, 200})
		}
	}

	// Colossal blade: + min(size*2, 10)
	if ctx.wearing("Colossal blade") {
		bonus := ctx.Monster.Size * 2
		if bonus > 10 {
			bonus = 10
		}
		maxHit += bonus
	}

	// Fang: shrink the hit range
	if ctx.wearing("Osmumten's fang") || ctx.wearing("Osmumten's fang (or)") {
		shrink := maxHit * 3 / 20
		return shrink, maxHit - shrink
	}

	return 0, maxHit
}

func (ctx *CalcContext) getPlayerMaxRangedHit() (int, int) {
	effectiveLevel := ctx.Stats.Ranged + ctx.Stats.RngBoost

	// Prayer
	if ctx.Prayer.FactorStrength[0] != 0 {
		effectiveLevel = ApplyFactor(effectiveLevel, ctx.Prayer.FactorStrength)
	}

	effectiveLevel += 8

	// Elite ranged void: * 9/8. Regular: * 11/10
	if ctx.isWearingEliteRangedVoid() {
		effectiveLevel = effectiveLevel * 9 / 8
	} else if ctx.isWearingRangedVoid() {
		effectiveLevel = effectiveLevel * 11 / 10
	}

	maxHit := MaxHitFromEffective(effectiveLevel, ctx.Bonuses.RangedStr+64)

	// Crystal bow/Bowfa: * (40 + crystalPieces) / 40
	if ctx.wearing("Crystal bow") || ctx.wearingAny("Bow of faerdhinen") {
		pieces := ctx.countCrystalPieces()
		maxHit = maxHit * (40 + pieces) / 40
	}

	// Salve/slayer (same pattern as melee)
	if ctx.wearing("Salve amulet(ei)") && ctx.monsterHasAttribute("undead") {
		maxHit = ApplyFactor(maxHit, Factor{6, 5})
	} else if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Black mask (i)")) && ctx.isOnSlayerTask() {
		maxHit = ApplyFactor(maxHit, Factor{23, 20})
	}

	// Twisted bow
	if ctx.wearing("Twisted bow") {
		maxHit = tbowScaling(maxHit, ctx.Monster, false)
	}

	// Dragon hunter crossbow + dragon
	if ctx.wearing("Dragon hunter crossbow") && ctx.monsterHasAttribute("dragon") {
		maxHit = ApplyFactor(maxHit, Factor{5, 4})
	}

	return 0, maxHit
}

func (ctx *CalcContext) getPlayerMaxMagicHit() (int, int) {
	magicLvl := ctx.Stats.Magic

	// Base max hit depends on weapon
	maxHit := ctx.poweredStaffMaxHit(magicLvl)
	if maxHit == 0 {
		return 0, 0
	}

	// Magic damage bonus
	// magicDmgBonus is in tenths of percent (e.g., 150 = 15.0%)
	magicDmgBonus := ctx.Bonuses.MagicStr

	// Prayer magic damage bonus
	magicDmgBonus += ctx.Prayer.MagicDmgBonus

	// Salve amulet(ei) + undead: +200 (20.0%)
	if ctx.wearing("Salve amulet(ei)") && ctx.monsterHasAttribute("undead") {
		magicDmgBonus += 200
	}

	// Smoke battlestaff + standard spellbook: +100 (10.0%)
	if ctx.wearing("Smoke battlestaff") || ctx.wearing("Mystic smoke staff") {
		magicDmgBonus += 100
	}

	// Apply magic damage bonus: maxHit = trunc(maxHit + trunc(maxHit * bonus / 1000))
	maxHit = maxHit + maxHit*magicDmgBonus/1000

	// Black mask (i) / slayer helm (i) for magic: * 23/20
	if (ctx.wearing("Slayer helmet (i)") || ctx.wearing("Black mask (i)")) && ctx.isOnSlayerTask() {
		maxHit = maxHit * 23 / 20
	}

	// Tome of fire + fire spell: * 11/10
	if ctx.wearing("Tome of fire") {
		maxHit = maxHit * 11 / 10
	}

	return 0, maxHit
}

func (ctx *CalcContext) poweredStaffMaxHit(magicLevel int) int {
	if ctx.Weapon == nil {
		return 0
	}
	name := ctx.Weapon.Name
	switch {
	case strings.Contains(name, "Sanguinesti staff"), strings.Contains(name, "Holy sanguinesti staff"):
		return max(1, magicLevel/3-1)
	case strings.Contains(name, "Trident of the swamp"):
		return max(1, magicLevel/3-2)
	case strings.Contains(name, "Trident of the seas"):
		return max(1, magicLevel/3-5)
	case strings.Contains(name, "Tumeken's shadow"):
		return max(1, magicLevel/3+1)
	case strings.Contains(name, "Warped sceptre"):
		return max(1, (8*magicLevel+96)/37)
	case strings.Contains(name, "Accursed sceptre"):
		return max(1, magicLevel/3-6)
	case strings.Contains(name, "Thammaron's sceptre"):
		return max(1, magicLevel/3-8)
	case strings.Contains(name, "Dawnbringer"):
		return max(1, magicLevel/6-1)
	case strings.Contains(name, "Starter staff"):
		return 8
	default:
		// For spell-based casting, caller should provide the spell max hit
		// For now, return 0 to indicate "not a powered staff"
		return 0
	}
}

// ── Twisted Bow Scaling ────────────────────────────────────────────

// tbowScaling matches the TypeScript tbowScaling function.
// accuracyMode=true for attack roll, false for max hit.
func tbowScaling(current int, monster *data.Monster, accuracyMode bool) int {
	factor := 14
	base := 250
	if accuracyMode {
		factor = 10
		base = 140
	}

	// Magic value: max(skills.magic, offensive.magic), capped at 250
	magic := monster.Skills.Magic
	if monster.Offensive.Magic > magic {
		magic = monster.Offensive.Magic
	}
	cap := 250
	if magic > cap {
		magic = cap
	}

	t2 := (3*magic - factor) / 100
	t3inner := 3*magic/10 - 10*factor
	t3 := (t3inner * t3inner) / 100

	bonus := base + t2 - t3
	return current * bonus / 100
}

// ── Scythe of Vitur ────────────────────────────────────────────────

func scytheHitsplats(monsterSize int) int {
	hits := monsterSize
	if hits < 1 {
		hits = 1
	}
	if hits > 3 {
		hits = 3
	}
	return hits
}

// ── Equipment check helpers ────────────────────────────────────────

func (ctx *CalcContext) wearing(item string) bool {
	lower := strings.ToLower(item)
	for _, equipped := range ctx.AllItems {
		if equipped == lower {
			return true
		}
	}
	return false
}

func (ctx *CalcContext) wearingAny(substring string) bool {
	lower := strings.ToLower(substring)
	for _, equipped := range ctx.AllItems {
		if strings.Contains(equipped, lower) {
			return true
		}
	}
	return false
}

func (ctx *CalcContext) monsterHasAttribute(attr string) bool {
	lower := strings.ToLower(attr)
	for _, a := range ctx.Monster.Attributes {
		if strings.ToLower(a) == lower {
			return true
		}
	}
	return false
}

func (ctx *CalcContext) isOnSlayerTask() bool {
	// CLI doesn't have slayer task context — assume true if slayer helm is worn
	return ctx.wearing("Slayer helmet") || ctx.wearing("Slayer helmet (i)") ||
		ctx.wearing("Black mask") || ctx.wearing("Black mask (i)")
}

func (ctx *CalcContext) isWearingMeleeVoid() bool {
	return ctx.wearingAny("void") && ctx.wearing("Void melee helm")
}

func (ctx *CalcContext) isWearingRangedVoid() bool {
	return ctx.wearingAny("void") && ctx.wearing("Void ranger helm")
}

func (ctx *CalcContext) isWearingEliteRangedVoid() bool {
	return ctx.wearing("Elite void top") && ctx.wearing("Elite void robe") &&
		ctx.wearing("Void knight gloves") && ctx.wearing("Void ranger helm")
}

func (ctx *CalcContext) isWearingMagicVoid() bool {
	return ctx.wearingAny("void") && ctx.wearing("Void mage helm")
}

func (ctx *CalcContext) countInquisitorPieces() int {
	count := 0
	if ctx.wearing("Inquisitor's great helm") {
		count++
	}
	if ctx.wearing("Inquisitor's hauberk") {
		count++
	}
	if ctx.wearing("Inquisitor's plateskirt") {
		count++
	}
	return count
}

func (ctx *CalcContext) isWearingFang() bool {
	return ctx.wearing("Osmumten's fang") || ctx.wearing("Osmumten's fang (or)")
}

func (ctx *CalcContext) isRevWeaponApplicable() bool {
	// Rev weapons need wilderness + charged version
	// CLI doesn't track wilderness state — return false by default
	// Users can override via preset
	return false
}

func (ctx *CalcContext) isTwoHitWeapon() bool {
	return ctx.wearing("Torag's hammers") || ctx.wearing("Sulphur blades") ||
		ctx.wearing("Glacial temotli") || ctx.wearing("Earthbound tecpatl")
}

func (ctx *CalcContext) isWearingKarils() bool {
	return ctx.wearing("Karil's crossbow") && ctx.wearing("Karil's coif") &&
		ctx.wearing("Karil's leathertop") && ctx.wearing("Karil's leatherskirt") &&
		ctx.wearing("Amulet of the damned")
}

func (ctx *CalcContext) isWearingAhrims() bool {
	return ctx.wearing("Ahrim's staff") && ctx.wearing("Ahrim's hood") &&
		ctx.wearing("Ahrim's robetop") && ctx.wearing("Ahrim's robeskirt") &&
		ctx.wearing("Amulet of the damned")
}

func (ctx *CalcContext) isWearingBloodMoon() bool {
	return ctx.wearing("Dual macuahuitl") && ctx.wearing("Blood moon helm") &&
		ctx.wearing("Blood moon chestplate") && ctx.wearing("Blood moon tassets")
}

func (ctx *CalcContext) isWearingBerserkerNecklace() bool {
	return ctx.wearing("Berserker necklace") || ctx.wearing("Berserker necklace (or)")
}

func (ctx *CalcContext) isWearingTzhaarWeapon() bool {
	return ctx.wearing("Tzhaar-ket-em") || ctx.wearing("Tzhaar-ket-om") ||
		ctx.wearing("Tzhaar-ket-om (t)") || ctx.wearing("Toktz-xil-ak") ||
		ctx.wearing("Toktz-xil-ek") || ctx.wearing("Toktz-mej-tal")
}

func (ctx *CalcContext) isWearingDharok() bool {
	return ctx.wearing("Dharok's greataxe") && ctx.wearing("Dharok's helm") &&
		ctx.wearing("Dharok's platebody") && ctx.wearing("Dharok's platelegs")
}

func (ctx *CalcContext) isWearingVeracs() bool {
	return ctx.wearing("Verac's flail") && ctx.wearing("Verac's helm") &&
		ctx.wearing("Verac's brassard") && ctx.wearing("Verac's plateskirt")
}

func (ctx *CalcContext) countCrystalPieces() int {
	count := 0
	if ctx.wearing("Crystal helm") {
		count++
	}
	if ctx.wearingAny("crystal legs") {
		count += 2
	}
	if ctx.wearingAny("crystal body") {
		count += 3
	}
	return count
}

// ── Gear resolution ────────────────────────────────────────────────

func resolveGearBonuses(gear GearSet) (TotalEquipmentBonuses, *data.Equipment, error) {
	var bonuses TotalEquipmentBonuses
	var weapon *data.Equipment

	for slot, itemName := range gear.GearSlots() {
		if itemName == "" {
			continue
		}
		equip, err := data.FindEquipment(itemName)
		if err != nil {
			continue // skip items not in dataset
		}

		bonuses.StabAtk += equip.Offensive.Stab
		bonuses.SlashAtk += equip.Offensive.Slash
		bonuses.CrushAtk += equip.Offensive.Crush
		bonuses.MagicAtk += equip.Offensive.Magic
		bonuses.RangedAtk += equip.Offensive.Ranged
		bonuses.MeleeStr += equip.Bonuses.Str
		bonuses.RangedStr += equip.Bonuses.RangedStr
		bonuses.MagicStr += equip.Bonuses.MagicStr

		if slot == "weapon" {
			weapon = equip
		}
	}

	return bonuses, weapon, nil
}

func determineCombatParams(gear GearSet, weapon *data.Equipment) (CombatStyle, PrayerData) {
	switch gear.Style {
	case "ranged":
		return StyleRanged, Prayers["rigour"]
	case "magic":
		return StyleMagic, Prayers["augury"]
	default:
		style := bestMeleeStyle(weapon)
		return style, Prayers["piety"]
	}
}

func bestMeleeStyle(weapon *data.Equipment) CombatStyle {
	if weapon == nil {
		return StyleCrush
	}
	cat := strings.ToLower(weapon.Category)
	switch {
	case strings.Contains(cat, "slash"):
		return StyleSlash
	case strings.Contains(cat, "stab"):
		return StyleStab
	case strings.Contains(cat, "crush"), strings.Contains(cat, "blunt"):
		return StyleCrush
	default:
		bestVal := weapon.Offensive.Slash
		style := StyleSlash
		if weapon.Offensive.Stab > bestVal {
			bestVal = weapon.Offensive.Stab
			style = StyleStab
		}
		if weapon.Offensive.Crush > bestVal {
			style = StyleCrush
		}
		return style
	}
}

// ── Formatting ─────────────────────────────────────────────────────

func formatTTK(ticks int) string {
	seconds := float64(ticks) * SecondsPerTick
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("~%d:%02d", mins, secs)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
