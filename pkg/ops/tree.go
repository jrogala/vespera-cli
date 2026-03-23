package ops

import (
	"fmt"
	"strings"

	"github.com/jrogala/vespera-cli/client"
)

// TreeEntry represents a node in the FTP directory tree.
type TreeEntry struct {
	Name     string      `json:"name"`
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size,omitempty"`
	Children []TreeEntry `json:"children,omitempty"`
}

// GetTree returns the directory tree starting at the given path.
func GetTree(c *client.FTPClient, rootPath string) ([]TreeEntry, error) {
	return walkTree(c.Conn(), rootPath)
}

func walkTree(conn client.FTPConn, path string) ([]TreeEntry, error) {
	entries, err := conn.List(path)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", path, err)
	}

	var result []TreeEntry
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		if e.Type == client.EntryTypeFolder {
			childPath := strings.TrimRight(path, "/") + "/" + e.Name
			children, _ := walkTree(conn, childPath)
			result = append(result, TreeEntry{
				Name:     e.Name,
				IsDir:    true,
				Children: children,
			})
		} else {
			result = append(result, TreeEntry{
				Name:  e.Name,
				IsDir: false,
				Size:  int64(e.Size),
			})
		}
	}
	return result, nil
}
