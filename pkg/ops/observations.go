package ops

import (
	"github.com/jrogala/vespera-cli/client"
)

// ListObservations returns all observation entries from the telescope.
func ListObservations(c *client.FTPClient) ([]client.ObservationEntry, error) {
	return c.ListObservations()
}

// ListFiles returns all files in a given observation.
func ListFiles(c *client.FTPClient, observation string) ([]client.FileEntry, error) {
	return c.ListFiles(observation)
}
