package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPEntry represents a directory entry from the FTP server.
type FTPEntry struct {
	Name  string
	Size  uint64
	Type  int // 0=file, 1=folder
	Time  time.Time
}

const (
	EntryTypeFile   = 0
	EntryTypeFolder = 1
)

// FTPConn is the interface for FTP operations, allowing mock implementations.
type FTPConn interface {
	Login(user, password string) error
	List(path string) ([]*FTPEntry, error)
	Retr(path string) (io.ReadCloser, error)
	Quit() error
}

// realFTPConn wraps the jlaffaye/ftp ServerConn to implement FTPConn.
type realFTPConn struct {
	conn *ftp.ServerConn
}

func (r *realFTPConn) Login(user, password string) error {
	return r.conn.Login(user, password)
}

func (r *realFTPConn) List(path string) ([]*FTPEntry, error) {
	entries, err := r.conn.List(path)
	if err != nil {
		return nil, err
	}
	var result []*FTPEntry
	for _, e := range entries {
		entryType := EntryTypeFile
		if e.Type == ftp.EntryTypeFolder {
			entryType = EntryTypeFolder
		}
		result = append(result, &FTPEntry{
			Name: e.Name,
			Size: e.Size,
			Type: entryType,
			Time: e.Time,
		})
	}
	return result, nil
}

func (r *realFTPConn) Retr(path string) (io.ReadCloser, error) {
	resp, err := r.conn.Retr(path)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *realFTPConn) Quit() error {
	return r.conn.Quit()
}

// DialFTP dials a real FTP server and returns an FTPConn.
func DialFTP(host string, port int) (FTPConn, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(10*time.Second))
	if err != nil {
		return nil, err
	}
	return &realFTPConn{conn: conn}, nil
}

// FTPClient handles file operations on the Vespera telescope via FTP.
type FTPClient struct {
	host string
	port int
	conn FTPConn
}

// NewFTPClient creates a new FTP client for the Vespera.
func NewFTPClient(host string, port int) *FTPClient {
	return &FTPClient{host: host, port: port}
}

// NewFTPClientWithConn creates an FTP client with an injected connection (for testing).
func NewFTPClientWithConn(host string, port int, conn FTPConn) *FTPClient {
	return &FTPClient{host: host, port: port, conn: conn}
}

// Connect establishes the FTP connection.
func (c *FTPClient) Connect() error {
	conn, err := DialFTP(c.host, c.port)
	if err != nil {
		return fmt.Errorf("connecting to %s:%d: %w", c.host, c.port, err)
	}
	if err := conn.Login("anonymous", ""); err != nil {
		conn.Quit()
		return fmt.Errorf("FTP login: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the FTP connection.
func (c *FTPClient) Close() {
	if c.conn != nil {
		c.conn.Quit()
	}
}

// Host returns the host address.
func (c *FTPClient) Host() string {
	return c.host
}

// Port returns the port.
func (c *FTPClient) Port() int {
	return c.port
}

// Conn returns the underlying FTP connection.
func (c *FTPClient) Conn() FTPConn {
	return c.conn
}

// ObservationEntry represents an observation folder on the Vespera.
type ObservationEntry struct {
	Name  string `json:"name"`
	Date  string `json:"date"`
	IsDir bool   `json:"is_dir"`
}

// ListObservations lists observation folders on the Vespera from both user and system.
func (c *FTPClient) ListObservations() ([]ObservationEntry, error) {
	var obs []ObservationEntry

	for _, dir := range []string{"/user", "/system/captures"} {
		entries, err := c.conn.List(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.Type == EntryTypeFolder && e.Name != "." && e.Name != ".." {
				obs = append(obs, ObservationEntry{
					Name:  e.Name,
					Date:  e.Time.Format("2006-01-02 15:04"),
					IsDir: true,
				})
			}
		}
	}
	return obs, nil
}

// FileEntry represents a file on the Vespera.
type FileEntry struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type string `json:"type"`
}

// ListFiles lists files recursively in a specific observation folder.
// Searches in /user/ first, then /system/captures/.
func (c *FTPClient) ListFiles(observation string) ([]FileEntry, error) {
	files, err := c.listFilesRecursive("/user/" + observation)
	if err == nil && len(files) > 0 {
		return files, nil
	}
	return c.listFilesRecursive("/system/captures/" + observation)
}

func (c *FTPClient) listFilesRecursive(path string) ([]FileEntry, error) {
	entries, err := c.conn.List(path)
	if err != nil {
		return nil, fmt.Errorf("listing %s: %w", path, err)
	}

	var files []FileEntry
	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		if e.Type == EntryTypeFolder {
			subFiles, err := c.listFilesRecursive(path + "/" + e.Name)
			if err != nil {
				continue
			}
			files = append(files, subFiles...)
		} else {
			files = append(files, FileEntry{
				Name: path + "/" + e.Name,
				Size: int64(e.Size),
				Type: FileType(e.Name),
			})
		}
	}
	return files, nil
}

// DownloadFile downloads a single file from the Vespera.
func (c *FTPClient) DownloadFile(remotePath, localDir string, fileSize int64) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	resp, err := c.conn.Retr(remotePath)
	if err != nil {
		return fmt.Errorf("retrieving %s: %w", remotePath, err)
	}
	defer resp.Close()

	filename := filepath.Base(remotePath)
	localPath := filepath.Join(localDir, filename)

	f, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp)
	return err
}

// DownloadObservation downloads all files from an observation folder.
func (c *FTPClient) DownloadObservation(observation, localDir string, filter string, workers int) (int, error) {
	files, err := c.ListFiles(observation)
	if err != nil {
		return 0, err
	}

	targetDir := filepath.Join(localDir, observation)
	count := 0

	// Calculate total
	var totalSize int64
	var filtered []FileEntry
	for _, f := range files {
		if filter != "" && !strings.EqualFold(f.Type, filter) {
			continue
		}
		filtered = append(filtered, f)
		totalSize += f.Size
	}

	fmt.Printf("Downloading %d files (%s)\n", len(filtered), FormatSize(totalSize))

	// Skip already-downloaded files (resume support)
	var toDownload []FileEntry
	var skipped int
	for _, f := range filtered {
		filename := filepath.Base(f.Name)
		localPath := filepath.Join(targetDir, filename)
		if info, err := os.Stat(localPath); err == nil && info.Size() == f.Size {
			skipped++
			continue
		}
		toDownload = append(toDownload, f)
	}
	if skipped > 0 {
		fmt.Printf("Skipping %d already downloaded files\n", skipped)
	}
	if len(toDownload) == 0 {
		fmt.Println("All files already downloaded")
		return len(filtered), nil
	}
	fmt.Printf("Downloading %d remaining files\n", len(toDownload))

	// Close the listing connection to free up FTP slots
	c.conn.Quit()
	c.conn = nil

	// Parallel download with multiple FTP connections
	if workers <= 0 {
		workers = 8
	}
	if len(toDownload) < workers {
		workers = len(toDownload)
	}

	total := len(toDownload)
	var done int64

	jobs := make(chan FileEntry, len(toDownload))
	results := make(chan error, len(toDownload))

	for w := 0; w < workers; w++ {
		go func() {
			wc := NewFTPClient(c.host, c.port)
			if err := wc.Connect(); err != nil {
				for range jobs {
					results <- err
				}
				return
			}
			defer wc.Close()
			for f := range jobs {
				err := wc.DownloadFile(f.Name, targetDir, f.Size)
				n := atomic.AddInt64(&done, 1)
				fmt.Printf("\r  %d/%d (%.0f%%)", n, total, float64(n)/float64(total)*100)
				results <- err
			}
		}()
	}

	for _, f := range toDownload {
		jobs <- f
	}
	close(jobs)

	for range toDownload {
		if err := <-results; err != nil {
			fmt.Println()
			return count + skipped, err
		}
		count++
	}
	fmt.Println()

	return count + skipped, nil
}

// FileType determines the file type from extension.
func FileType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".fits", ".fit":
		return "FITS"
	case ".tiff", ".tif":
		return "TIFF"
	case ".jpg", ".jpeg":
		return "JPEG"
	case ".png":
		return "PNG"
	default:
		return ext
	}
}

// FormatSize formats bytes to human-readable.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
