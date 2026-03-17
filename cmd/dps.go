package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
	"github.com/JordanCoin/osrs-wiki-cli/internal/dps"
	"github.com/spf13/cobra"
)

var dpsCmd = &cobra.Command{
	Use:   "dps",
	Short: "Calculate DPS against a monster",
	Long: `Calculate damage-per-second for a gear preset against any OSRS monster.
Uses the same formulas as the OSRS Wiki DPS calculator.

Available presets: max-melee, max-range, max-mage, mid-melee, mid-range, void-range

Use --weapon to override just the weapon in any preset.
Use --compare to compare multiple weapons side by side.`,
	Example: `  osrs-wiki dps --preset max-melee --monster "Vardorvis"
  osrs-wiki dps --preset max-melee --monster "Vardorvis" --version "Awakened"
  osrs-wiki dps --preset max-melee --weapon "Ghrazi rapier" --monster "Vardorvis"
  osrs-wiki dps --compare "Scythe of vitur,Ghrazi rapier" --preset max-melee --monster "Vardorvis"
  osrs-wiki dps --preset max-range --monster "Vorkath"
  osrs-wiki dps --presets`,
	RunE: runDPS,
}

func runDPS(cmd *cobra.Command, args []string) error {
	listPresets, _ := cmd.Flags().GetBool("presets")
	if listPresets {
		fmt.Println("Available presets:")
		names := dps.PresetNames()
		sort.Strings(names)
		for _, name := range names {
			preset := dps.Presets[name]
			fmt.Printf("  %-12s  %s (weapon: %s)\n", name, preset.Name, preset.Weapon)
		}
		return nil
	}

	presetName, _ := cmd.Flags().GetString("preset")
	monsterName, _ := cmd.Flags().GetString("monster")
	monsterVersion, _ := cmd.Flags().GetString("version")
	weaponOverride, _ := cmd.Flags().GetString("weapon")
	compareStr, _ := cmd.Flags().GetString("compare")

	if presetName == "" {
		fmt.Fprintln(os.Stderr, "Error: --preset is required. Use --presets to list available presets.")
		os.Exit(1)
	}
	if monsterName == "" {
		fmt.Fprintln(os.Stderr, "Error: --monster is required.")
		os.Exit(1)
	}

	preset, ok := dps.Presets[presetName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: unknown preset '%s'. Use --presets to list available presets.\n", presetName)
		os.Exit(1)
	}

	monster, err := data.FindMonsterVersion(monsterName, monsterVersion)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(3)
	}

	stats := dps.MaxedStats()

	// Compare mode
	if compareStr != "" {
		weapons := strings.Split(compareStr, ",")
		for i := range weapons {
			weapons[i] = strings.TrimSpace(weapons[i])
		}
		return runCompare(preset, weapons, monster, stats)
	}

	// Single weapon mode
	if weaponOverride != "" {
		preset.Weapon = weaponOverride
		// If overriding with a 1h weapon on a 2h preset, keep the shield empty
		// User can further customize if needed
	}

	result, err := dps.Calculate(preset, monster, stats)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if jsonOutput {
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))
		return nil
	}

	printDPSResult(result, monster)
	return nil
}

func runCompare(preset dps.GearSet, weapons []string, monster *data.Monster, stats dps.PlayerStats) error {
	type compareEntry struct {
		Weapon string
		Result *dps.DPSResult
	}

	var results []compareEntry
	for _, w := range weapons {
		p := preset
		p.Weapon = w
		result, err := dps.Calculate(p, monster, stats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping '%s': %s\n", w, err)
			continue
		}
		results = append(results, compareEntry{w, result})
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no valid weapons to compare")
		os.Exit(1)
	}

	if jsonOutput {
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))
		return nil
	}

	fmt.Printf("Comparison at %s (%s):\n", monster.DisplayName(), preset.Name)

	// Sort by DPS descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Result.DPS > results[j].Result.DPS
	})

	for _, r := range results {
		fmt.Printf("  %-25s  DPS %-6.2f  TTK %s\n", r.Result.Weapon, r.Result.DPS, r.Result.TTKString)
	}

	if len(results) >= 2 {
		best := results[0]
		second := results[1]
		pctBetter := (best.Result.DPS - second.Result.DPS) / second.Result.DPS * 100
		fmt.Printf("  Winner: %s (+%.1f%%)\n", best.Result.Weapon, pctBetter)
	}

	return nil
}

func printDPSResult(result *dps.DPSResult, monster *data.Monster) {
	fmt.Printf("# DPS: %s vs %s\n", result.Weapon, monster.DisplayName())
	fmt.Printf("  Weapon: %s | Style: %s | Prayer: %s\n", result.Weapon, result.Style, result.Prayer)
	fmt.Printf("  Max hit: %d | Accuracy: %.1f%% | DPS: %.2f\n", result.MaxHit, result.Accuracy, result.DPS)
	fmt.Printf("  TTK: %s (%d ticks)\n", result.TTKString, result.TTKTicks)
}

func init() {
	dpsCmd.Flags().String("preset", "", "Equipment preset (max-melee, max-range, max-mage, mid-melee, mid-range, void-range)")
	dpsCmd.Flags().String("monster", "", "Monster name to calculate DPS against")
	dpsCmd.Flags().String("version", "", "Specific monster version, e.g. Post-quest or Awakened")
	dpsCmd.Flags().String("weapon", "", "Override the preset weapon")
	dpsCmd.Flags().String("compare", "", "Comma-separated weapon names to compare")
	dpsCmd.Flags().Bool("presets", false, "List all available presets")
}
