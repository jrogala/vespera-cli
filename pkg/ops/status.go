package ops

import (
	"github.com/jrogala/vespera-cli/client"
)

// StatusResult holds telescope connection status.
type StatusResult struct {
	Host             string `json:"host"`
	ObservationCount int    `json:"observation_count"`
}

// GetStatus checks telescope connectivity and returns status info.
func GetStatus(c *client.FTPClient) (*StatusResult, error) {
	obs, err := c.ListObservations()
	if err != nil {
		return nil, err
	}
	return &StatusResult{
		Host:             c.Host(),
		ObservationCount: len(obs),
	}, nil
}
