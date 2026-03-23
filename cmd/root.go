// Package cmd implements the vespera-cli commands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jrogala/vespera-cli/config"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "vespera",
	Short: "CLI for Vaonis Vespera 2 telescope",
	Long: `CLI for controlling the Vaonis Vespera 2 telescope and processing images.

The telescope runs on its own WiFi (vespera2-XXXX). Use --auto-wifi to auto-connect,
run the command, and reconnect to your home WiFi. Use --timeout to set max seconds.

Quick examples:
  vespera list --auto-wifi                           List observations (auto WiFi switch)
  vespera files <observation> --auto-wifi            List files in an observation
  vespera download <observation> --auto-wifi         Download all files
  vespera download <obs> --type FITS --auto-wifi --timeout 600  Download FITS (10min timeout)
  vespera tree --auto-wifi                           Show directory tree on telescope
  vespera status --auto-wifi                         Check telescope connection`,
}

func Execute() {
	config.Init()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Output raw JSON responses")
	rootCmd.SetHelpFunc(customHelp)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

func customHelp(cmd *cobra.Command, _ []string) {
	if cmd == rootCmd {
		printTree()
		return
	}
	if !cmd.HasSubCommands() {
		printLeafHelp(cmd)
		return
	}
	printSubtree(cmd)
}

func printTree() {
	fmt.Println("vespera - Vaonis Vespera 2 telescope CLI")
	fmt.Println("")
	fmt.Println("Global: --json (raw JSON output)")
	fmt.Println("")
	fmt.Println("Commands:")

	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		if cmd.HasSubCommands() {
			fmt.Printf("  %s\n", cmd.Name())
			for _, sub := range cmd.Commands() {
				if sub.Hidden {
					continue
				}
				aliases := ""
				if len(sub.Aliases) > 0 {
					aliases = " (" + strings.Join(sub.Aliases, ", ") + ")"
				}
				fmt.Printf("    %-10s %s%s\n", sub.Name(), sub.Short, aliases)
			}
		} else {
			aliases := ""
			if len(cmd.Aliases) > 0 {
				aliases = " (" + strings.Join(cmd.Aliases, ", ") + ")"
			}
			fmt.Printf("  %-12s %s%s\n", cmd.Name(), cmd.Short, aliases)
		}
	}

	fmt.Println("")
	fmt.Println("Run 'vespera <command> --help' for full details.")
}

func printSubtree(cmd *cobra.Command) {
	fmt.Printf("%s\n\n", cmd.Short)

	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		aliases := ""
		if len(sub.Aliases) > 0 {
			aliases = " (" + strings.Join(sub.Aliases, ", ") + ")"
		}
		fmt.Printf("  %-10s %s%s\n", sub.Name(), sub.Short, aliases)
	}

	fmt.Println("")
	fmt.Printf("Run 'vespera %s <subcommand> --help' for full details.\n", cmd.Name())
}

func printLeafHelp(cmd *cobra.Command) {
	fmt.Printf("%s %s\n", cmd.UseLine(), "")
	fmt.Println(cmd.Short)

	if cmd.HasLocalFlags() {
		fmt.Println("")
		fmt.Println("Flags:")
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			shorthand := ""
			if f.Shorthand != "" {
				shorthand = "-" + f.Shorthand + ", "
			}
			def := ""
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
				def = " (default: " + f.DefValue + ")"
			}
			fmt.Printf("  %s--%s %s%s\n", shorthand, f.Name, f.Usage, def)
		})
	}
}
