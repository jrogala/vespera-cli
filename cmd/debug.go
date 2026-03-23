package cmd

import (
	"fmt"

	"github.com/jrogala/vespera-cli/client"
	"github.com/jrogala/vespera-cli/internal/cmdutil"
	"github.com/jrogala/vespera-cli/pkg/ops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(treeCmd)
	treeCmd.Flags().Bool("auto-wifi", false, "Auto-connect to Vespera WiFi")
	treeCmd.Flags().Int("timeout", 300, "Max seconds on Vespera WiFi")
}

var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Show full directory tree on telescope",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "/"
		if len(args) > 0 {
			root = "/" + args[0]
		}
		return wrapAutoWifi(cmd, func() error {
			c, err := cmdutil.NewClient()
			if err != nil {
				return err
			}
			defer c.Close()

			tree, err := ops.GetTree(c, root)
			if err != nil {
				return err
			}

			cmdutil.Render(cmd, tree, func() {
				printTreeEntries(tree, "")
			})
			return nil
		})
	},
}

func printTreeEntries(entries []ops.TreeEntry, indent string) {
	for _, e := range entries {
		if e.IsDir {
			fmt.Printf("%s%s/\n", indent, e.Name)
			printTreeEntries(e.Children, indent+"  ")
		} else {
			fmt.Printf("%s%s (%s)\n", indent, e.Name, client.FormatSize(e.Size))
		}
	}
}
