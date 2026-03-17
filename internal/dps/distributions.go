package dps

// Hit distributions for special attacks and multi-hit weapons.
// Ported from weirdgloop/osrs-dps-calc HitDist.ts + PlayerVsNPCCalc.ts

// ── Dragon Claws Special Attack ────────────────────────────────────

// DragonClawsSpec computes the expected damage for dragon claws spec.
// 4 accuracy rolls, each successively tries if previous failed.
// Roll 0 hits: split as [dmg/2, dmg/4, dmg/8, dmg/8+1]
// Roll 1 hits: [dmg/2, dmg/4, dmg/4+1, 0]
// Roll 2 hits: [dmg/2, dmg/2+1, 0, 0]
// Roll 3 hits: [dmg+1, 0, 0, 0]
// All miss: 2/3 chance [1,1,0,0], 1/3 chance [0,0,0,0]
func DragonClawsExpectedDmg(accuracy float64, maxHit int) float64 {
	miss := 1.0 - accuracy

	// Roll 0 hits (accuracy)
	r0Avg := float64(0)
	for dmg := 0; dmg <= maxHit; dmg++ {
		total := dmg/2 + dmg/4 + dmg/8 + (dmg/8 + 1)
		r0Avg += float64(total)
	}
	r0Avg /= float64(maxHit + 1)

	// Roll 1 hits (miss * accuracy)
	r1Avg := float64(0)
	for dmg := 0; dmg <= maxHit; dmg++ {
		low := maxHit * 3 / 4
		high := maxHit + low - 1
		if dmg < low || dmg > high {
			continue
		}
		total := dmg/2 + dmg/4 + (dmg/4 + 1)
		r1Avg += float64(total)
	}
	r1Range := maxHit*1/4 + 1
	if r1Range > 0 {
		r1Avg /= float64(r1Range)
	}

	// Roll 2 hits (miss^2 * accuracy)
	r2Avg := float64(0)
	for dmg := 0; dmg <= maxHit; dmg++ {
		low := maxHit * 2 / 4
		high := maxHit + low - 1
		if dmg < low || dmg > high {
			continue
		}
		total := dmg/2 + (dmg/2 + 1)
		r2Avg += float64(total)
	}
	r2Range := maxHit*2/4 + 1
	if r2Range > 0 {
		r2Avg /= float64(r2Range)
	}

	// Roll 3 hits (miss^3 * accuracy)
	r3Avg := float64(maxHit+1) / 2.0

	// All miss: 2/3 chance [1,1], 1/3 chance [0,0]
	missAvg := 2.0 / 3.0 * 2.0

	expected := accuracy*r0Avg +
		miss*accuracy*r1Avg +
		miss*miss*accuracy*r2Avg +
		miss*miss*miss*accuracy*r3Avg +
		miss*miss*miss*miss*missAvg

	return expected
}

// ── Burning Claws Special Attack ───────────────────────────────────

// BurningClawsExpectedDmg computes expected damage for burning claws spec.
// 3 accuracy rolls with split damage + DoT burn chance.
func BurningClawsExpectedDmg(accuracy float64, maxHit int) float64 {
	miss := 1.0 - accuracy

	// Roll 0: [dmg/2, dmg/4, dmg/4]
	r0Avg := float64(0)
	for dmg := 0; dmg <= maxHit; dmg++ {
		total := dmg/2 + dmg/4 + dmg/4
		r0Avg += float64(total)
	}
	r0Avg /= float64(maxHit + 1)

	// Roll 1: [dmg/2-1, dmg/2-1, 2]
	r1Avg := float64(0)
	r1Count := 0
	for dmg := 0; dmg <= maxHit; dmg++ {
		low := maxHit * 2 / 4
		high := maxHit + low
		if dmg < low || dmg > high {
			continue
		}
		h1 := dmg/2 - 1
		if h1 < 0 {
			h1 = 0
		}
		total := h1 + h1 + 2
		r1Avg += float64(total)
		r1Count++
	}
	if r1Count > 0 {
		r1Avg /= float64(r1Count)
	}

	// Roll 2: [dmg-2, 1, 1]
	r2Avg := float64(0)
	r2Count := 0
	for dmg := 0; dmg <= maxHit; dmg++ {
		low := maxHit * 1 / 4
		high := maxHit + low
		if dmg < low || dmg > high {
			continue
		}
		h0 := dmg - 2
		if h0 < 0 {
			h0 = 0
		}
		total := h0 + 1 + 1
		r2Avg += float64(total)
		r2Count++
	}
	if r2Count > 0 {
		r2Avg /= float64(r2Count)
	}

	// All miss: 1/5 [0], 2/5 [1], 2/5 [1,1]
	missAvg := 2.0/5.0*1.0 + 2.0/5.0*2.0

	// Burn DoT: 0.15 per hitsplat, 10 damage each
	burnExpected := 0.15 * 3.0 * 10.0 * accuracy // rough average

	expected := accuracy*r0Avg +
		miss*accuracy*r1Avg +
		miss*miss*accuracy*r2Avg +
		miss*miss*miss*missAvg +
		burnExpected

	return expected
}

// ── Dark Bow ───────────────────────────────────────────────────────

// DarkBowExpectedDmg computes expected damage for dark bow (2 arrows).
// During spec: dragon arrows min=8, *15/10. Other arrows min=5, *13/10.
func DarkBowExpectedDmg(accuracy float64, maxHit int, dragonArrows bool) float64 {
	specMult := 13
	minHit := 5
	if dragonArrows {
		specMult = 15
		minHit = 8
	}

	specMax := maxHit * specMult / 10
	avgPerHit := accuracy * float64(specMax+minHit) / 2.0
	return avgPerHit * 2.0 // two arrows
}

// ── Tonalztics of Ralos ────────────────────────────────────────────

// TonalzticsExpectedDmg computes expected damage for tonalztics (charged).
// Two hits, each at 75% max.
func TonalzticsExpectedDmg(accuracy float64, maxHit int) float64 {
	reducedMax := maxHit * 3 / 4
	return accuracy * float64(reducedMax) / 2.0 * 2.0
}

// ── Abyssal Dagger Special ─────────────────────────────────────────

// AbyssalDaggerExpectedDmg: first hit rolls accuracy. If accurate,
// second hit has 100% accuracy. If inaccurate, second also inaccurate.
func AbyssalDaggerExpectedDmg(accuracy float64, maxHit int) float64 {
	// Max hit is reduced to 85% (* 17/20)
	specMax := maxHit * 17 / 20
	hitAvg := float64(specMax) / 2.0

	// Accurate: both hits land
	// Inaccurate: both miss
	return accuracy * hitAvg * 2.0
}

// ── Saradomin Sword Special ────────────────────────────────────────

// SaradomSwordExpectedDmg: melee hit + magic hit (1-16).
// Magic hit only on accurate melee hit.
func SaradomSwordExpectedDmg(accuracy float64, meleeMax int) float64 {
	meleeAvg := accuracy * float64(meleeMax) / 2.0
	magicAvg := accuracy * 8.5 // average of 1-16
	return meleeAvg + magicAvg
}

// ── Granite Hammer Special ─────────────────────────────────────────

// GraniteHammerExpectedDmg: +5 flat to every hit (including inaccurate).
func GraniteHammerExpectedDmg(accuracy float64, maxHit int) float64 {
	return accuracy*float64(maxHit)/2.0 + 5.0
}

// ── P2 Wardens Damage Modifier ─────────────────────────────────────

// P2WardensModifier computes the min/max hit range for P2 Wardens.
// Accuracy is always 1.0.
func P2WardensModifier(maxHit, atkRoll, defRoll int) (int, int) {
	reducedDef := defRoll / 3
	accuracyDelta := atkRoll - reducedDef
	if accuracyDelta < 0 {
		accuracyDelta = 0
	}

	modifier := iLerp(15, 40, 0, 42000, accuracyDelta)
	if modifier < 15 {
		modifier = 15
	}
	if modifier > 40 {
		modifier = 40
	}

	minHit := maxHit * modifier / 100
	newMax := maxHit * (modifier + 20) / 100
	return minHit, newMax
}

// iLerp: integer linear interpolation.
// Returns x1 + trunc((x2-x1) * (yc-y1) / (y2-y1))
func iLerp(x1, x2, y1, y2, yc int) int {
	if y2 == y1 {
		return x1
	}
	return x1 + (x2-x1)*(yc-y1)/(y2-y1)
}

// ── Gadderhammer ───────────────────────────────────────────────────

// GadderhammerExpectedDmg: vs shades, 95% chance *5/4, 5% chance *2.
func GadderhammerExpectedDmg(accuracy float64, maxHit int, isShade bool) float64 {
	if !isShade {
		return accuracy * float64(maxHit) / 2.0
	}
	avg54 := float64(maxHit*5/4) / 2.0
	avg2x := float64(maxHit*2) / 2.0
	return accuracy * (0.95*avg54 + 0.05*avg2x)
}

// ── Two-hit weapons ────────────────────────────────────────────────

// TwoHitExpectedDmg: Torag's hammers, Sulphur blades, etc.
// Two independent hits: first max = max/2, second max = max - max/2.
func TwoHitExpectedDmg(accuracy float64, maxHit int) float64 {
	h1Max := maxHit / 2
	h2Max := maxHit - h1Max
	return accuracy * (float64(h1Max)/2.0 + float64(h2Max)/2.0)
}

// ── Dual Macuahuitl ────────────────────────────────────────────────

// DualMacuahuitlExpectedDmg: first hit max = max/2, second = max - max/2.
// Second hit only rolls if first is accurate.
func DualMacuahuitlExpectedDmg(accuracy float64, maxHit int) float64 {
	h1Max := maxHit / 2
	h2Max := maxHit - h1Max
	// Second hit only on accurate first hit
	return accuracy*float64(h1Max)/2.0 + accuracy*accuracy*float64(h2Max)/2.0
}

// ── Karil's Set (Amulet of Damned) ─────────────────────────────────

// KarilsExpectedDmg: 75% normal, 25% adds second splat at trunc(first/2).
// Only on accurate hits.
func KarilsExpectedDmg(accuracy float64, maxHit int) float64 {
	normalAvg := float64(maxHit) / 2.0
	procAvg := float64(maxHit)/2.0 + float64(maxHit/2)/2.0 // original + half
	return accuracy * (0.75*normalAvg + 0.25*procAvg)
}

// ── Ahrim's Set (Amulet of Damned) ─────────────────────────────────

// AhrimsExpectedDmg: 75% normal, 25% damage * 13/10.
func AhrimsExpectedDmg(accuracy float64, maxHit int) float64 {
	normalAvg := float64(maxHit) / 2.0
	procAvg := float64(maxHit*13/10) / 2.0
	return accuracy * (0.75*normalAvg + 0.25*procAvg)
}

// ── Blood Moon Set ─────────────────────────────────────────────────

// BloodMoonExpectedDmg: minHit = max/4, maxHit = max + max/4.
func BloodMoonExpectedDmg(accuracy float64, maxHit int) float64 {
	minHit := maxHit / 4
	newMax := maxHit + minHit
	return accuracy * float64(newMax+minHit) / 2.0
}

// ── Voidwaker Special ──────────────────────────────────────────────

// VoidwakerExpectedDmg: min = max/2, max = max + max/2. Always 100% accuracy.
func VoidwakerExpectedDmg(maxHit int) float64 {
	minHit := maxHit / 2
	newMax := maxHit + minHit
	return float64(newMax+minHit) / 2.0 // 100% accuracy
}

// ── Abyssal Bludgeon Special ───────────────────────────────────────

// AbyssalBludgeonExpectedDmg: * (100 + missingPrayer/2) / 100.
func AbyssalBludgeonExpectedDmg(accuracy float64, maxHit int, prayer, maxPrayer int) float64 {
	missing := maxPrayer - prayer
	if missing < 0 {
		missing = 0
	}
	specMax := maxHit * (100 + missing/2) / 100
	return accuracy * float64(specMax) / 2.0
}

// ── Soulreaper Axe ─────────────────────────────────────────────────

// SoulreaperAxeExpectedDmg: accuracy and damage scale with stacks.
// * (100 + 6*stacks) / 100 for both.
func SoulreaperAxeExpectedDmg(accuracy float64, maxHit, atkRoll int, stacks int) float64 {
	factor := Factor{100 + 6*stacks, 100}
	specMax := ApplyFactor(maxHit, factor)
	specAtk := ApplyFactor(atkRoll, factor)
	_ = specAtk // accuracy recalculated but we simplify here
	return accuracy * float64(specMax) / 2.0
}

// ── Rev Weapons ────────────────────────────────────────────────────

// RevWeaponFactor returns the attack/damage multiplier for rev weapons
// when in wilderness and charged.
func RevWeaponFactor() Factor {
	return Factor{3, 2}
}

// ── Chinchompa Distance Scaling ────────────────────────────────────

// ChinchompaAccuracyFactor returns the accuracy factor based on fuse and distance.
func ChinchompaAccuracyFactor(fuse string, distance int) Factor {
	// Short fuse: 4/3/2 for dist 1-3/4-6/7
	// Medium fuse: 3/4/3
	// Long fuse: 2/3/4
	var num int
	switch fuse {
	case "short":
		switch {
		case distance <= 3:
			num = 4
		case distance <= 6:
			num = 3
		default:
			num = 2
		}
	case "medium":
		switch {
		case distance <= 3:
			num = 3
		case distance <= 6:
			num = 4
		default:
			num = 3
		}
	case "long":
		switch {
		case distance <= 3:
			num = 2
		case distance <= 6:
			num = 3
		default:
			num = 4
		}
	default:
		num = 4
	}
	return Factor{num, 4}
}
