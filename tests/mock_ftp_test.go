package tests

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jrogala/vespera-cli/client"
)

// mockFTPEntry is a directory/file entry in the mock FTP.
type mockFTPEntry struct {
	name    string
	isDir   bool
	size    uint64
	time    time.Time
	content string // file content for Retr
}

// mockFTPConn implements client.FTPConn for testing.
type mockFTPConn struct {
	// filesystem maps directory path -> list of entries in that directory
	filesystem map[string][]mockFTPEntry
	loggedIn   bool
}

func newMockFTPConn() *mockFTPConn {
	return &mockFTPConn{
		filesystem: make(map[string][]mockFTPEntry),
	}
}

func (m *mockFTPConn) Login(user, password string) error {
	m.loggedIn = true
	return nil
}

func (m *mockFTPConn) List(path string) ([]*client.FTPEntry, error) {
	entries, ok := m.filesystem[path]
	if !ok {
		return nil, fmt.Errorf("no such directory: %s", path)
	}
	var result []*client.FTPEntry
	for _, e := range entries {
		entryType := client.EntryTypeFile
		if e.isDir {
			entryType = client.EntryTypeFolder
		}
		result = append(result, &client.FTPEntry{
			Name: e.name,
			Size: e.size,
			Type: entryType,
			Time: e.time,
		})
	}
	return result, nil
}

func (m *mockFTPConn) Retr(path string) (io.ReadCloser, error) {
	// Find the file in the filesystem
	dir := path[:strings.LastIndex(path, "/")]
	if dir == "" {
		dir = "/"
	}
	base := path[strings.LastIndex(path, "/")+1:]

	entries, ok := m.filesystem[dir]
	if !ok {
		return nil, fmt.Errorf("no such directory: %s", dir)
	}
	for _, e := range entries {
		if e.name == base && !e.isDir {
			return io.NopCloser(strings.NewReader(e.content)), nil
		}
	}
	return nil, fmt.Errorf("no such file: %s", path)
}

func (m *mockFTPConn) Quit() error {
	return nil
}

// addDir adds a directory entry to the mock filesystem.
func (m *mockFTPConn) addDir(parentPath, name string, t time.Time) {
	m.filesystem[parentPath] = append(m.filesystem[parentPath], mockFTPEntry{
		name:  name,
		isDir: true,
		time:  t,
	})
}

// addFile adds a file entry to the mock filesystem.
func (m *mockFTPConn) addFile(parentPath, name string, size uint64, content string) {
	m.filesystem[parentPath] = append(m.filesystem[parentPath], mockFTPEntry{
		name:    name,
		isDir:   false,
		size:    size,
		content: content,
	})
}

// ensureDir ensures a directory path exists in the filesystem (creates empty listing).
func (m *mockFTPConn) ensureDir(path string) {
	if _, ok := m.filesystem[path]; !ok {
		m.filesystem[path] = []mockFTPEntry{}
	}
}
