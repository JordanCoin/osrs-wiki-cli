package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/JordanCoin/osrs-wiki-cli/internal/data"
	"github.com/spf13/cobra"
)

var monsterCmd = &cobra.Command{
	Use:   "monster [name]",
	Short: "Look up monster combat stats",
	Long: `Look up a monster's combat stats from the OSRS DPS calculator dataset.
Shows HP, defence levels, defence bonuses, weaknesses, and attack style.
Data is downloaded on first use and cached locally.`,
	Args: cobra.MinimumNArgs(1),
	Example: `  osrs-wiki monster "Vardorvis"
  osrs-wiki monster "Vorkath"
  osrs-wiki monster "General Graardor"
  osrs-wiki monster "Vardorvis" --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")

		monster, err := data.FindMonster(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(3)
		}

		if jsonOutput {
			out, _ := json.MarshalIndent(monster, "", "  ")
			fmt.Println(string(out))
			return nil
		}

		fmt.Printf("# %s\n", monster.DisplayName())
		fmt.Printf("  Combat level: %d\n", monster.Level)
		fmt.Printf("  HP: %d\n", monster.Skills.HP)
		fmt.Printf("  Size: %dx%d\n", monster.Size, monster.Size)
		fmt.Printf("  Attack speed: %d ticks (%.1fs)\n", monster.Speed, float64(monster.Speed)*0.6)

		if len(monster.Style) > 0 {
			fmt.Printf("  Attack style: %s\n", strings.Join(monster.Style, ", "))
		}
		if monster.MaxHit != "" && monster.MaxHit != "0" {
			fmt.Printf("  Max hit: %s\n", monster.MaxHit)
		}

		fmt.Println("\n  Skills:")
		fmt.Printf("    ATK: %-4d  STR: %-4d  DEF: %d\n", monster.Skills.Atk, monster.Skills.Str, monster.Skills.Def)
		fmt.Printf("    Magic: %-4d  Ranged: %d\n", monster.Skills.Magic, monster.Skills.Ranged)

		fmt.Println("\n  Defence bonuses:")
		fmt.Printf("    Stab: %-4d  Slash: %-4d  Crush: %d\n", monster.Defensive.Stab, monster.Defensive.Slash, monster.Defensive.Crush)
		fmt.Printf("    Magic: %-4d  Ranged (std): %d\n", monster.Defensive.Magic, monster.Defensive.Standard)
		if monster.Defensive.Light != 0 || monster.Defensive.Heavy != 0 {
			fmt.Printf("    Ranged (light): %-4d  Ranged (heavy): %d\n", monster.Defensive.Light, monster.Defensive.Heavy)
		}

		if monster.Weakness != nil {
			fmt.Printf("\n  Weakness: %s (%d%%)\n", monster.Weakness.Element, monster.Weakness.Severity)
		}

		if len(monster.Attributes) > 0 {
			fmt.Printf("  Attributes: %s\n", strings.Join(monster.Attributes, ", "))
		}

		return nil
	},
}
