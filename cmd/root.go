package cmd

import "github.com/spf13/cobra"

var jsonOutput bool
var appVersion string

func SetVersion(v string) { appVersion = v }

var rootCmd = &cobra.Command{
	Use:   "osrs-wiki",
	Short: "OSRS Wiki CLI — item images, prices, and search",
	Long: `Look up Old School RuneScape items on the OSRS Wiki.
Get correct image URLs, Grand Exchange prices, and item details.

Agent-friendly: supports --json output, meaningful exit codes, no interactive prompts.`,
}

func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	rootCmd.AddCommand(imageCmd)
	rootCmd.AddCommand(priceCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(itemCmd)
	rootCmd.AddCommand(wikiCmd)
	rootCmd.AddCommand(guideCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run:   func(cmd *cobra.Command, args []string) { cmd.Println("osrs-wiki " + appVersion) },
	})
}
