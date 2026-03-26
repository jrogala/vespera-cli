package ops

import (
	"github.com/jrogala/vespera-cli/client"
)

// DownloadOptions configures a download operation.
type DownloadOptions struct {
	Observation string
	OutputDir   string
	TypeFilter  string
	Workers     int
}

// DownloadResult holds download operation results.
type DownloadResult struct {
	Observation string `json:"observation"`
	FileCount   int    `json:"file_count"`
	OutputDir   string `json:"output_dir"`
}

// DownloadObservation downloads files from an observation folder.
func DownloadObservation(c *client.FTPClient, opts DownloadOptions) (*DownloadResult, error) {
	count, err := c.DownloadObservation(opts.Observation, opts.OutputDir, opts.TypeFilter, opts.Workers)
	if err != nil {
		return nil, err
	}
	return &DownloadResult{
		Observation: opts.Observation,
		FileCount:   count,
		OutputDir:   opts.OutputDir,
	}, nil
}
