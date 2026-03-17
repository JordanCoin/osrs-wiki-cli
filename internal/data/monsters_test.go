package data

import (
	"encoding/json"
	"testing"
)

func TestMonsterMaxHitUnmarshalSupportsNumbersAndStrings(t *testing.T) {
	t.Run("number", func(t *testing.T) {
		var monster Monster
		if err := json.Unmarshal([]byte(`{"max_hit":0}`), &monster); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if monster.MaxHit != "0" {
			t.Fatalf("got %q, want %q", monster.MaxHit, "0")
		}
	})

	t.Run("string", func(t *testing.T) {
		var monster Monster
		if err := json.Unmarshal([]byte(`{"max_hit":"30 (Magic)"}`), &monster); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if monster.MaxHit != "30 (Magic)" {
			t.Fatalf("got %q, want %q", monster.MaxHit, "30 (Magic)")
		}
	})

	t.Run("control-characters", func(t *testing.T) {
		var monster Monster
		if err := json.Unmarshal([]byte("{\"max_hit\":\"32-43 (Melee)\\u007f\\u007f\"}"), &monster); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if monster.MaxHit != "32-43 (Melee)" {
			t.Fatalf("got %q, want %q", monster.MaxHit, "32-43 (Melee)")
		}
	})
}

func TestPickBestMonsterMatchPrefersRepeatableDefaults(t *testing.T) {
	matches := []Monster{
		{Name: "Vardorvis", Version: "Awakened", Skills: struct {
			Atk    int `json:"atk"`
			Def    int `json:"def"`
			HP     int `json:"hp"`
			Magic  int `json:"magic"`
			Ranged int `json:"ranged"`
			Str    int `json:"str"`
		}{HP: 1400}},
		{Name: "Vardorvis", Version: "Post-quest", Skills: struct {
			Atk    int `json:"atk"`
			Def    int `json:"def"`
			HP     int `json:"hp"`
			Magic  int `json:"magic"`
			Ranged int `json:"ranged"`
			Str    int `json:"str"`
		}{HP: 700}},
		{Name: "Vardorvis", Version: "Quest", Skills: struct {
			Atk    int `json:"atk"`
			Def    int `json:"def"`
			HP     int `json:"hp"`
			Magic  int `json:"magic"`
			Ranged int `json:"ranged"`
			Str    int `json:"str"`
		}{HP: 500}},
	}

	best := pickBestMonsterMatch(matches)
	if best.Version != "Post-quest" {
		t.Fatalf("got version %q, want %q", best.Version, "Post-quest")
	}
}
