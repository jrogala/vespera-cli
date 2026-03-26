package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/jrogala/vespera-cli/client"
	"github.com/jrogala/vespera-cli/config"
	"github.com/jrogala/vespera-cli/internal/cmdutil"
	"github.com/jrogala/vespera-cli/pkg/ops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd, listCmd, filesCmd, downloadCmd)

	// Add --auto-wifi and --timeout to all FTP commands
	for _, cmd := range []*cobra.Command{statusCmd, listCmd, filesCmd, downloadCmd} {
		cmd.Flags().Bool("auto-wifi", false, "Auto-connect to Vespera WiFi, run command, reconnect to previous WiFi")
		cmd.Flags().Int("timeout", 300, "Max seconds on Vespera WiFi before force-reconnect (default 5min)")
	}

	downloadCmd.Flags().String("type", "", "Filter by file type (FITS, TIFF, JPEG)")
	downloadCmd.Flags().StringP("output", "o", "", "Output directory (default: ~/Pictures/vespera)")
	downloadCmd.Flags().IntP("workers", "w", 8, "Number of parallel download workers")
}

// wrapAutoWifi wraps a function with auto WiFi connect/disconnect if --auto-wifi is set.
func wrapAutoWifi(cmd *cobra.Command, fn func() error) error {
	autoWifi, _ := cmd.Flags().GetBool("auto-wifi")
	if autoWifi {
		timeout, _ := cmd.Flags().GetInt("timeout")
		return client.WithVesperaTimeout(fn, time.Duration(timeout)*time.Second)
	}
	return fn()
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check telescope connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		return wrapAutoWifi(cmd, func() error {
			c, err := cmdutil.NewClient()
			if err != nil {
				return err
			}
			defer c.Close()

			result, err := ops.GetStatus(c)
			if err != nil {
				return err
			}

			cmdutil.Render(cmd, result, func() {
				fmt.Printf("Connected to Vespera at %s\n", result.Host)
				fmt.Printf("Observations: %d\n", result.ObservationCount)
			})
			return nil
		})
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List observations on telescope",
	RunE: func(cmd *cobra.Command, args []string) error {
		return wrapAutoWifi(cmd, func() error {
			c, err := cmdutil.NewClient()
			if err != nil {
				return err
			}
			defer c.Close()

			entries, err := ops.ListObservations(c)
			if err != nil {
				return err
			}

			cmdutil.Render(cmd, entries, func() {
				if len(entries) == 0 {
					fmt.Println("No observations found")
					return
				}
				w := cmdutil.NewTabWriter()
				fmt.Fprintln(w, "NAME\tDATE")
				for _, o := range entries {
					fmt.Fprintf(w, "%s\t%s\n", o.Name, o.Date)
				}
				w.Flush()
			})
			return nil
		})
	},
}

var filesCmd = &cobra.Command{
	Use:   "files <observation>",
	Short: "List files in an observation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return wrapAutoWifi(cmd, func() error {
			c, err := cmdutil.NewClient()
			if err != nil {
				return err
			}
			defer c.Close()

			files, err := ops.ListFiles(c, args[0])
			if err != nil {
				return err
			}

			cmdutil.Render(cmd, files, func() {
				if len(files) == 0 {
					fmt.Println("No files found")
					return
				}
				w := cmdutil.NewTabWriter()
				fmt.Fprintln(w, "NAME\tTYPE\tSIZE")
				for _, f := range files {
					fmt.Fprintf(w, "%s\t%s\t%s\n", f.Name, f.Type, client.FormatSize(f.Size))
				}
				w.Flush()
			})
			return nil
		})
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <observation>",
	Short: "Download files from an observation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return wrapAutoWifi(cmd, func() error {
			c, err := cmdutil.NewClient()
			if err != nil {
				return err
			}
			defer c.Close()

			filter, _ := cmd.Flags().GetString("type")
			outputDir, _ := cmd.Flags().GetString("output")
			if outputDir == "" {
				outputDir = config.OutputDir()
			}
			workers, _ := cmd.Flags().GetInt("workers")

			fmt.Printf("Downloading %s to %s...\n", args[0], filepath.Join(outputDir, args[0]))
			result, err := ops.DownloadObservation(c, ops.DownloadOptions{
				Observation: args[0],
				OutputDir:   outputDir,
				TypeFilter:  filter,
				Workers:     workers,
			})
			if err != nil {
				return err
			}

			cmdutil.Render(cmd, result, func() {
				fmt.Printf("Downloaded %d files\n", result.FileCount)
			})
			return nil
		})
	},
}
