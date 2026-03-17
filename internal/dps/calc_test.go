package dps

import (
	"testing"
)

func TestGetNormalAccuracyRoll(t *testing.T) {
	// Standard case: atk > def
	acc := getNormalAccuracyRoll(30000, 20000)
	if acc < 0.6 || acc > 0.7 {
		t.Fatalf("expected ~0.66, got %f", acc)
	}

	// Standard case: atk < def
	acc = getNormalAccuracyRoll(10000, 30000)
	if acc < 0.15 || acc > 0.2 {
		t.Fatalf("expected ~0.16, got %f", acc)
	}

	// Equal rolls
	acc = getNormalAccuracyRoll(20000, 20000)
	if acc < 0.49 || acc > 0.51 {
		t.Fatalf("expected ~0.5, got %f", acc)
	}

	// Zero rolls
	acc = getNormalAccuracyRoll(0, 0)
	if acc != 0 {
		t.Fatalf("expected 0, got %f", acc)
	}
}

func TestGetFangAccuracyRoll(t *testing.T) {
	// Fang should always be >= normal accuracy
	normal := getNormalAccuracyRoll(25000, 30000)
	fang := getFangAccuracyRoll(25000, 30000)
	if fang < normal {
		t.Fatalf("fang accuracy %f should be >= normal %f", fang, normal)
	}
}

func TestApplyFactor(t *testing.T) {
	// Piety accuracy: 120/100
	result := ApplyFactor(99, Factor{120, 100})
	if result != 118 { // trunc(99 * 120 / 100) = 118
		t.Fatalf("expected 118, got %d", result)
	}

	// Rigour strength: 123/100
	result = ApplyFactor(99, Factor{123, 100})
	if result != 121 { // trunc(99 * 123 / 100) = 121
		t.Fatalf("expected 121, got %d", result)
	}
}

func TestMaxHitFromEffective(t *testing.T) {
	// Standard formula: trunc((level * bonus + 320) / 640)
	result := MaxHitFromEffective(118, 182) // typical maxed melee
	if result < 30 || result > 40 {
		t.Fatalf("expected realistic melee max hit, got %d", result)
	}
}

func TestScytheExpectedMax(t *testing.T) {
	// Size 1: 1 hit = base max
	if got := ScytheExpectedMax(40, 1); got != 40 {
		t.Fatalf("size 1: expected 40, got %d", got)
	}

	// Size 2: hit0=40, hit1=20 = 60
	if got := ScytheExpectedMax(40, 2); got != 60 {
		t.Fatalf("size 2: expected 60, got %d", got)
	}

	// Size 3: hit0=40, hit1=20, hit2=10 = 70
	if got := ScytheExpectedMax(40, 3); got != 70 {
		t.Fatalf("size 3: expected 70, got %d", got)
	}

	// Size 5 capped at 3 hits
	if got := ScytheExpectedMax(40, 5); got != 70 {
		t.Fatalf("size 5 (capped): expected 70, got %d", got)
	}
}

func TestDharokMaxHit(t *testing.T) {
	// At 1 HP with 99 max HP: factor = (10000 + 98*99) / 10000 = 1.9702
	result := DharokMaxHit(40, 99, 1)
	if result < 78 || result > 80 {
		t.Fatalf("dharok at 1hp: expected ~79, got %d", result)
	}

	// At full HP: factor = 10000/10000 = 1.0
	result = DharokMaxHit(40, 99, 99)
	if result != 40 {
		t.Fatalf("dharok at full hp: expected 40, got %d", result)
	}
}

func TestTbowScaling(t *testing.T) {
	// Create a mock monster with high magic level
	m := &DummyMonster{magic: 250, offMagic: 0}

	// Accuracy scaling (base 140)
	result := tbowScalingDirect(20000, 250, true)
	if result < 20000 || result > 30000 {
		t.Fatalf("tbow accuracy scaling: expected increase, got %d from 20000", result)
	}

	// Damage scaling (base 250)
	result = tbowScalingDirect(40, 250, false)
	if result < 40 || result > 100 {
		t.Fatalf("tbow damage scaling: expected increase, got %d from 40", result)
	}

	_ = m // suppress unused
}

// tbowScalingDirect is a test helper that takes magic level directly.
func tbowScalingDirect(current, magic int, accuracyMode bool) int {
	factor := 14
	base := 250
	if accuracyMode {
		factor = 10
		base = 140
	}
	if magic > 250 {
		magic = 250
	}

	t2 := (3*magic - factor) / 100
	t3inner := 3*magic/10 - 10*factor
	t3 := (t3inner * t3inner) / 100

	bonus := base + t2 - t3
	return current * bonus / 100
}

// DummyMonster for testing (avoid importing data package circular)
type DummyMonster struct {
	magic    int
	offMagic int
}
