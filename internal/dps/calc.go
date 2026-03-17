package dps

import (
	"fmt"
	"math"
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

// CombatStyle represents an attack style for DPS calculation.
type CombatStyle string

const (
	StyleStab   CombatStyle = "Stab"
	StyleSlash  CombatStyle = "Slash"
	StyleCrush  CombatStyle = "Crush"
	StyleRanged CombatStyle = "Ranged"
	StyleMagic  CombatStyle = "Magic"
)

// Prayer bonuses for top-tier prayers.
type PrayerBonus struct {
	Name    string
	AtkMult float64
	StrMult float64
}

var (
	Piety   = PrayerBonus{"Piety", 1.20, 1.23}
	Rigour  = PrayerBonus{"Rigour", 1.20, 1.23}
	Augury  = PrayerBonus{"Augury", 1.25, 0} // magic doesn't boost str via prayer the same way
	NoPrayer = PrayerBonus{"None", 1.0, 1.0}
)

// PlayerStats represents the player's combat levels.
type PlayerStats struct {
	Attack   int
	Strength int
	Defence  int
	Ranged   int
	Magic    int
	Prayer   int
	Hitpoints int
}

// MaxedStats returns maxed player combat stats.
func MaxedStats() PlayerStats {
	return PlayerStats{
		Attack:    99,
		Strength:  99,
		Defence:   99,
		Ranged:    99,
		Magic:     99,
		Prayer:    99,
		Hitpoints: 99,
	}
}

// TotalEquipmentBonuses sums up bonuses from all equipped items.
type TotalEquipmentBonuses struct {
	StabAtk   int
	SlashAtk  int
	CrushAtk  int
	MagicAtk  int
	RangedAtk int
	MeleeStr  int
	RangedStr int
	MagicStr  int
}

// DPSResult holds the output of a DPS calculation.
type DPSResult struct {
	Weapon    string      `json:"weapon"`
	Style     CombatStyle `json:"style"`
	Prayer    string      `json:"prayer"`
	MaxHit    int         `json:"max_hit"`
	Accuracy  float64     `json:"accuracy"`
	DPS       float64     `json:"dps"`
	TTKTicks  int         `json:"ttk_ticks"`
	TTKString string      `json:"ttk_string"`
	MonsterHP int         `json:"monster_hp"`
	AtkSpeed  int         `json:"attack_speed"`
}

// Calculate computes DPS for a gear set against a monster.
func Calculate(gear GearSet, monster *data.Monster, stats PlayerStats) (*DPSResult, error) {
	// Load all equipment pieces and sum bonuses
	bonuses, weapon, err := resolveGearBonuses(gear)
	if err != nil {
		return nil, err
	}

	if weapon == nil {
		return nil, fmt.Errorf("no weapon found in gear set")
	}

	// Determine combat style and prayer based on gear style
	style, prayer := determineCombatParams(gear, weapon)

	// Calculate attack and strength
	atkRoll := calcAttackRoll(stats, bonuses, style, prayer)
	defRoll := calcDefenceRoll(monster, style)
	accuracy := calcAccuracy(atkRoll, defRoll)
	maxHit := calcMaxHit(stats, bonuses, style, prayer)

	// Apply special weapon/gear modifiers
	maxHit, accuracy = applySpecialModifiers(maxHit, accuracy, weapon, monster, gear)

	atkSpeed := weapon.Speed
	if atkSpeed <= 0 {
		atkSpeed = 4
	}

	dps := calcDPS(accuracy, maxHit, atkSpeed)
	ttkTicks := calcTTK(monster.Skills.HP, dps, atkSpeed)
	ttkStr := formatTTK(ttkTicks)

	return &DPSResult{
		Weapon:    weapon.Name,
		Style:     style,
		Prayer:    prayer.Name,
		MaxHit:    maxHit,
		Accuracy:  math.Round(accuracy*1000) / 10, // percentage with 1 decimal
		DPS:       math.Round(dps*100) / 100,
		TTKTicks:  ttkTicks,
		TTKString: ttkStr,
		MonsterHP: monster.Skills.HP,
		AtkSpeed:  atkSpeed,
	}, nil
}

func resolveGearBonuses(gear GearSet) (TotalEquipmentBonuses, *data.Equipment, error) {
	var bonuses TotalEquipmentBonuses
	var weapon *data.Equipment

	for slot, itemName := range gear.GearSlots() {
		equip, err := data.FindEquipment(itemName)
		if err != nil {
			// Skip items we can't find (some preset items may not be in the dataset)
			continue
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

func determineCombatParams(gear GearSet, weapon *data.Equipment) (CombatStyle, PrayerBonus) {
	switch gear.Style {
	case "ranged":
		return StyleRanged, Rigour
	case "magic":
		return StyleMagic, Augury
	default:
		// For melee, pick the best style based on weapon category
		style := bestMeleeStyle(weapon)
		return style, Piety
	}
}

func bestMeleeStyle(weapon *data.Equipment) CombatStyle {
	cat := strings.ToLower(weapon.Category)
	switch {
	case strings.Contains(cat, "slash"):
		return StyleSlash
	case strings.Contains(cat, "stab"):
		return StyleStab
	case strings.Contains(cat, "crush"), strings.Contains(cat, "blunt"):
		return StyleCrush
	default:
		// Pick whichever offensive bonus is highest
		max := weapon.Offensive.Slash
		style := StyleSlash
		if weapon.Offensive.Stab > max {
			max = weapon.Offensive.Stab
			style = StyleStab
		}
		if weapon.Offensive.Crush > max {
			style = StyleCrush
		}
		return style
	}
}

func calcAttackRoll(stats PlayerStats, bonuses TotalEquipmentBonuses, style CombatStyle, prayer PrayerBonus) int {
	var baseLevel int
	var equipBonus int

	switch style {
	case StyleRanged:
		baseLevel = stats.Ranged
		equipBonus = bonuses.RangedAtk
	case StyleMagic:
		baseLevel = stats.Magic
		equipBonus = bonuses.MagicAtk
	case StyleStab:
		baseLevel = stats.Attack
		equipBonus = bonuses.StabAtk
	case StyleSlash:
		baseLevel = stats.Attack
		equipBonus = bonuses.SlashAtk
	case StyleCrush:
		baseLevel = stats.Attack
		equipBonus = bonuses.CrushAtk
	}

	// Effective level = floor(base * prayer) + stance bonus + 8
	effectiveLevel := int(math.Floor(float64(baseLevel) * prayer.AtkMult))
	effectiveLevel += 8 // potionless, stance=0 (accurate would add 3)

	return effectiveLevel * (equipBonus + 64)
}

func calcDefenceRoll(monster *data.Monster, style CombatStyle) int {
	var defBonus int

	switch style {
	case StyleStab:
		defBonus = monster.Defensive.Stab
	case StyleSlash:
		defBonus = monster.Defensive.Slash
	case StyleCrush:
		defBonus = monster.Defensive.Crush
	case StyleMagic:
		defBonus = monster.Defensive.Magic
	case StyleRanged:
		// Use "standard" ranged defence as default
		defBonus = monster.Defensive.Standard
	}

	// Monster effective defence = defence level + 9
	effectiveDefLevel := monster.Skills.Def + 9

	return effectiveDefLevel * (defBonus + 64)
}

func calcAccuracy(atkRoll, defRoll int) float64 {
	atk := float64(atkRoll)
	def := float64(defRoll)

	if atk > def {
		return 1.0 - (def+2.0)/(2.0*(atk+1.0))
	}
	return atk / (2.0*def + 1.0)
}

func calcMaxHit(stats PlayerStats, bonuses TotalEquipmentBonuses, style CombatStyle, prayer PrayerBonus) int {
	var baseLevel int
	var strBonus int

	switch style {
	case StyleRanged:
		baseLevel = stats.Ranged
		strBonus = bonuses.RangedStr
	case StyleMagic:
		// Magic max hit is different — based on spell/staff
		// For powered staves, use magic_str bonus
		baseLevel = stats.Magic
		strBonus = bonuses.MagicStr
	default:
		// Melee
		baseLevel = stats.Strength
		strBonus = bonuses.MeleeStr
	}

	effectiveStr := int(math.Floor(float64(baseLevel) * prayer.StrMult))
	effectiveStr += 8

	if style == StyleMagic {
		// Powered staves: base max hit = magic level / 3 - 1, scaled by magic_str
		// Simplified: use the standard formula but magic str is % based
		baseMax := int(math.Floor(float64(effectiveStr)*(float64(strBonus)+64.0)/640.0)) + 1
		// Apply magic damage bonus as percentage
		if bonuses.MagicStr > 0 {
			baseMax = int(math.Floor(float64(baseMax) * (1.0 + float64(bonuses.MagicStr)/100.0)))
		}
		return baseMax
	}

	return int(math.Floor(float64(effectiveStr)*(float64(strBonus)+64.0)/640.0)) + 1
}

func applySpecialModifiers(maxHit int, accuracy float64, weapon *data.Equipment, monster *data.Monster, gear GearSet) (int, float64) {
	name := strings.ToLower(weapon.Name)

	// Scythe of vitur: 3 hitsplats (100%, 50%, 25%) on large monsters
	// We handle this by increasing effective max hit
	if strings.Contains(name, "scythe of vitur") && monster.Size >= 2 {
		// Average across 3 hitsplats: max_hit * (1 + 0.5 + 0.25) = 1.75x effective
		maxHit = int(math.Floor(float64(maxHit) * 1.75))
	} else if strings.Contains(name, "scythe of vitur") {
		// Size 1: only first hit lands
		// Keep maxHit as is
	}

	// Twisted bow: accuracy and damage scale with monster's magic level
	if strings.Contains(name, "twisted bow") {
		magicLvl := monster.Skills.Magic
		if magicLvl > 250 {
			magicLvl = 250
		}
		// Tbow accuracy: 140 + (3*magic - 10) / 100 - (3*magic/10 - 100)^2 / 100
		tbowAcc := 140.0 + (3.0*float64(magicLvl)-10.0)/100.0 - math.Pow(3.0*float64(magicLvl)/10.0-100.0, 2)/100.0
		if tbowAcc > 140 {
			tbowAcc = 140
		}
		tbowDmg := 250.0 + (3.0*float64(magicLvl)-14.0)/100.0 - math.Pow(3.0*float64(magicLvl)/10.0-140.0, 2)/100.0
		if tbowDmg > 250 {
			tbowDmg = 250
		}
		accuracy = accuracy * tbowAcc / 100.0
		if accuracy > 1.0 {
			accuracy = 1.0
		}
		maxHit = int(math.Floor(float64(maxHit) * tbowDmg / 100.0))
	}

	// Void equipment bonus
	if isVoidSet(gear) {
		if gear.Style == "ranged" {
			accuracy *= 1.10
			maxHit = int(math.Floor(float64(maxHit) * 1.125))
		} else if gear.Style == "melee" {
			accuracy *= 1.10
			maxHit = int(math.Floor(float64(maxHit) * 1.10))
		}
	}

	// Salve amulet / slayer helm bonuses for undead/task — skip for simplicity

	return maxHit, accuracy
}

func isVoidSet(gear GearSet) bool {
	return strings.Contains(strings.ToLower(gear.Head), "void") &&
		strings.Contains(strings.ToLower(gear.Body), "void") &&
		strings.Contains(strings.ToLower(gear.Legs), "void") &&
		strings.Contains(strings.ToLower(gear.Hands), "void")
}

func calcDPS(accuracy float64, maxHit int, atkSpeed int) float64 {
	expectedHit := accuracy * float64(maxHit) / 2.0
	return expectedHit / (float64(atkSpeed) * 0.6)
}

func calcTTK(hp int, dps float64, atkSpeed int) int {
	if dps <= 0 {
		return 0
	}
	seconds := float64(hp) / dps
	ticks := int(math.Ceil(seconds / 0.6))
	// Round to nearest attack tick
	ticks = ((ticks + atkSpeed - 1) / atkSpeed) * atkSpeed
	return ticks
}

func formatTTK(ticks int) string {
	seconds := float64(ticks) * 0.6
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("~%d:%02d", mins, secs)
}
