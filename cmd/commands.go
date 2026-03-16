package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JordanCoin/osrs-wiki-cli/internal/api"
	"github.com/spf13/cobra"
)

// ── Image command ────────────────────────────────────────────────────

var imageCmd = &cobra.Command{
	Use:   "image [item name]",
	Short: "Get the correct OSRS Wiki image URL for an item",
	Long: `Looks up the exact thumbnail URL from the OSRS Wiki MediaWiki API.
Never guesses filenames — always returns the correct URL that renders.`,
	Args: cobra.ExactArgs(1),
	Example: `  osrs-wiki image "Twisted Bow"
  osrs-wiki image "Scythe of Vitur" --size 200
  osrs-wiki image "Dragon Claws" --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		size, _ := cmd.Flags().GetInt("size")
		full, _ := cmd.Flags().GetBool("full")
		client := api.NewClient()
		result, err := client.GetImage(args[0], size)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(3)
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
		} else if full {
			fmt.Println(result.FullURL)
		} else {
			fmt.Println(result.ImageURL)
		}
		return nil
	},
}

// ── Price command ────────────────────────────────────────────────────

var priceCmd = &cobra.Command{
	Use:     "price [item name]",
	Short:   "Get the Grand Exchange price for an item",
	Args:    cobra.ExactArgs(1),
	Example: `  osrs-wiki price "Twisted Bow"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient()
		result, err := client.GetPrice(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(3)
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Printf("%s: %s gp (buy: %s / sell: %s)\n",
				result.Name,
				formatGP(result.High),
				formatGP(result.High),
				formatGP(result.Low))
		}
		return nil
	},
}

// ── Search command ───────────────────────────────────────────────────

var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Short:   "Search for items on the OSRS Wiki",
	Args:    cobra.ExactArgs(1),
	Example: `  osrs-wiki search "dragon" --limit 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		client := api.NewClient()
		results, err := client.Search(args[0], limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(out))
		} else {
			if len(results) == 0 {
				fmt.Printf("No results for '%s'\n", args[0])
				os.Exit(3)
			}
			for _, r := range results {
				fmt.Printf("  %s\n", r.Title)
			}
		}
		return nil
	},
}

// ── Item command ─────────────────────────────────────────────────────

var itemCmd = &cobra.Command{
	Use:     "item [item name]",
	Short:   "Get item details (examine, alch values, members)",
	Args:    cobra.ExactArgs(1),
	Example: `  osrs-wiki item "Twisted Bow"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient()
		info, err := client.GetItem(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(3)
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(info, "", "  ")
			fmt.Println(string(out))
		} else {
			members := "No"
			if info.Members {
				members = "Yes"
			}
			fmt.Printf("%s (ID: %d)\n", info.Name, info.ID)
			fmt.Printf("  Examine: %s\n", info.Examine)
			fmt.Printf("  Members: %s\n", members)
			fmt.Printf("  High Alch: %s gp\n", formatGP(int64(info.HighAlch)))
			fmt.Printf("  Low Alch: %s gp\n", formatGP(int64(info.LowAlch)))
		}
		return nil
	},
}

func init() {
	imageCmd.Flags().Int("size", 150, "Thumbnail size in pixels")
	imageCmd.Flags().Bool("full", false, "Return full-resolution image URL instead of thumbnail")
	searchCmd.Flags().Int("limit", 10, "Max results")
}

func formatGP(n int64) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.0fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
