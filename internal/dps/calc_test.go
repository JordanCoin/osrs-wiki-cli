package dps

import (
	"testing"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
)

func TestCalcMaxHitUsesPoweredStaffFormula(t *testing.T) {
	weapon := &data.Equipment{Name: "Sanguinesti staff"}
	stats := MaxedStats()
	bonuses := TotalEquipmentBonuses{MagicStr: 24}

	maxHit := calcMaxHit(stats, bonuses, StyleMagic, Augury, weapon)
	if maxHit < 30 {
		t.Fatalf("got max hit %d, want a realistic powered-staff hit", maxHit)
	}
}

func TestScytheDamageMultiplierByMonsterSize(t *testing.T) {
	if got := scytheDamageMultiplier(1); got != 1.0 {
		t.Fatalf("size 1 multiplier = %v, want 1.0", got)
	}
	if got := scytheDamageMultiplier(2); got != 1.5 {
		t.Fatalf("size 2 multiplier = %v, want 1.5", got)
	}
	if got := scytheDamageMultiplier(3); got != 1.75 {
		t.Fatalf("size 3 multiplier = %v, want 1.75", got)
	}
}
