package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JordanCoin/osrs-wiki-cli/internal/api"
	"github.com/spf13/cobra"
)

// ── Wiki page command ────────────────────────────────────────────────

var wikiCmd = &cobra.Command{
	Use:   "wiki [page title]",
	Short: "Read any OSRS Wiki page content",
	Long: `Fetch the text content of any OSRS Wiki page. Returns clean plaintext
with section headings. Use this to look up bosses, items, quests, skills,
minigames, or any game mechanic instead of guessing from training data.`,
	Args: cobra.MinimumNArgs(1),
	Example: `  osrs-wiki wiki "Twisted Bow"
  osrs-wiki wiki "Vorkath/Strategies"
  osrs-wiki wiki "Chambers of Xeric"
  osrs-wiki wiki "Guardians of the Rift"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		// Join multiple args as title (so quotes aren't always needed)
		if len(args) > 1 {
			for _, a := range args[1:] {
				title += " " + a
			}
		}
		chars, _ := cmd.Flags().GetInt("chars")

		client := api.NewClient()
		page, err := client.GetPage(title, chars)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(3)
		}
		if jsonOutput {
			out, _ := json.MarshalIndent(page, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Printf("# %s\n", page.Title)
			fmt.Printf("Source: %s\n\n", page.URL)
			fmt.Println(page.Content)
		}
		return nil
	},
}

// ── Guide command — shortcuts for common guide pages ─────────────────

var guideCmd = &cobra.Command{
	Use:   "guide [topic]",
	Short: "Look up training guides, boss strategies, and money makers",
	Long: `Shortcuts for common OSRS Wiki guide pages:
  - Skill training: "osrs-wiki guide runecraft" → Runecraft training guide
  - Boss strategy: "osrs-wiki guide vorkath" → Vorkath/Strategies
  - Money making: "osrs-wiki guide money" → Money making guide
  - Quest: "osrs-wiki guide "Dragon Slayer II"" → Quest guide`,
	Args: cobra.MinimumNArgs(1),
	Example: `  osrs-wiki guide runecraft
  osrs-wiki guide vorkath
  osrs-wiki guide money
  osrs-wiki guide "Dragon Slayer II"
  osrs-wiki guide slayer`,
	RunE: func(cmd *cobra.Command, args []string) error {
		topic := args[0]
		if len(args) > 1 {
			for _, a := range args[1:] {
				topic += " " + a
			}
		}
		chars, _ := cmd.Flags().GetInt("chars")

		// Resolve common guide page names
		title := resolveGuideTitle(topic)

		client := api.NewClient()
		page, err := client.GetPage(title, chars)
		if err != nil {
			// Try with "_training" suffix for skills
			page, err = client.GetPage(topic+" training", chars)
			if err != nil {
				// Try as-is (maybe it's a direct page name)
				page, err = client.GetPage(topic, chars)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s\nTry 'osrs-wiki search %s' to find the right page.\n", err, topic)
					os.Exit(3)
				}
			}
		}

		if jsonOutput {
			out, _ := json.MarshalIndent(page, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Printf("# %s\n", page.Title)
			fmt.Printf("Source: %s\n\n", page.URL)
			fmt.Println(page.Content)
		}
		return nil
	},
}

// resolveGuideTitle maps common topics to wiki page titles.
func resolveGuideTitle(topic string) string {
	// Known boss strategy pages
	bossStrategies := map[string]string{
		"vorkath":             "Vorkath/Strategies",
		"zulrah":              "Zulrah/Strategies",
		"cerberus":            "Cerberus/Strategies",
		"grotesque guardians": "Grotesque Guardians/Strategies",
		"alchemical hydra":    "Alchemical Hydra/Strategies",
		"nightmare":           "The Nightmare/Strategies",
		"nex":                 "Nex/Strategies",
		"phantom muspah":      "Phantom Muspah/Strategies",
		"duke sucellus":       "Duke Sucellus/Strategies",
		"the leviathan":       "The Leviathan/Strategies",
		"the whisperer":       "The Whisperer/Strategies",
		"vardorvis":           "Vardorvis/Strategies",
		"tob":                 "Chambers of Xeric/Strategies",
		"cox":                 "Chambers of Xeric/Strategies",
		"toa":                 "Tombs of Amascut/Strategies",
	}

	// Known skill training pages (wiki uses "Pay-to-play X training")
	skillTraining := map[string]string{
		"runecraft":    "Pay-to-play Runecraft training",
		"runecrafting": "Pay-to-play Runecraft training",
		"slayer":       "Pay-to-play Slayer training",
		"mining":       "Pay-to-play Mining training",
		"smithing":     "Pay-to-play Smithing training",
		"fishing":      "Pay-to-play Fishing training",
		"cooking":      "Pay-to-play Cooking training",
		"woodcutting":  "Pay-to-play Woodcutting training",
		"firemaking":   "Pay-to-play Firemaking training",
		"agility":      "Pay-to-play Agility training",
		"thieving":     "Pay-to-play Thieving training",
		"farming":      "Pay-to-play Farming training",
		"herblore":     "Pay-to-play Herblore training",
		"crafting":     "Pay-to-play Crafting training",
		"fletching":    "Pay-to-play Fletching training",
		"hunter":       "Pay-to-play Hunter training",
		"construction": "Pay-to-play Construction training",
		"prayer":       "Pay-to-play Prayer training",
		"magic":        "Pay-to-play Magic training",
		"ranged":       "Pay-to-play Ranged training",
		"strength":     "Pay-to-play Strength training",
		"attack":       "Pay-to-play Attack training",
		"defence":      "Pay-to-play Defence training",
		"hitpoints":    "Pay-to-play Hitpoints training",
		"sailing":      "Sailing training",
	}

	// Special pages
	special := map[string]string{
		"money":        "Money making guide",
		"money making": "Money making guide",
		"moneymaking":  "Money making guide",
		"quests":       "Optimal quest guide",
		"quest order":  "Optimal quest guide",
	}

	lower := fmt.Sprintf("%s", topic)
	// Normalize to lowercase for lookup
	lowerTopic := lower

	if title, ok := bossStrategies[lowerTopic]; ok {
		return title
	}
	if title, ok := skillTraining[lowerTopic]; ok {
		return title
	}
	if title, ok := special[lowerTopic]; ok {
		return title
	}

	return topic
}

func init() {
	wikiCmd.Flags().Int("chars", 4000, "Max characters to fetch")
	guideCmd.Flags().Int("chars", 4000, "Max characters to fetch")
}
