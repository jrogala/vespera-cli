package client

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// WiFiManager handles auto-connect/disconnect to the Vespera hotspot.
type WiFiManager struct {
	previousSSID string
	vesperaSSID  string
	connected    bool
}

// NewWiFiManager creates a new WiFi manager.
func NewWiFiManager() *WiFiManager {
	return &WiFiManager{}
}

// CurrentSSID returns the currently connected WiFi SSID.
func (w *WiFiManager) CurrentSSID() (string, error) {
	cmd := exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi")
	cmd.Env = append(cmd.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting current WiFi: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "yes:") {
			return strings.TrimPrefix(line, "yes:"), nil
		}
	}
	return "", fmt.Errorf("no active WiFi connection found")
}

// FindVespera scans for Vespera WiFi networks.
func (w *WiFiManager) FindVespera() (string, error) {
	exec.Command("nmcli", "dev", "wifi", "rescan").Run()
	time.Sleep(2 * time.Second)

	out, err := exec.Command("nmcli", "-t", "-f", "ssid", "dev", "wifi", "list").Output()
	if err != nil {
		return "", fmt.Errorf("scanning WiFi: %w", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		ssid := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(ssid), "vespera") {
			return ssid, nil
		}
	}
	return "", fmt.Errorf("no Vespera WiFi found")
}

// Connect saves current WiFi, finds Vespera, and connects to it.
func (w *WiFiManager) Connect() error {
	current, err := w.CurrentSSID()
	if err != nil {
		return err
	}
	w.previousSSID = current

	ssid, err := w.FindVespera()
	if err != nil {
		return err
	}
	w.vesperaSSID = ssid

	fmt.Printf("Switching from %q to %q...\n", w.previousSSID, w.vesperaSSID)

	out, err := exec.Command("nmcli", "dev", "wifi", "connect", w.vesperaSSID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("connecting to %s: %s", w.vesperaSSID, string(out))
	}

	time.Sleep(3 * time.Second)
	w.connected = true
	return nil
}

// Disconnect reconnects to the previous WiFi network.
func (w *WiFiManager) Disconnect() error {
	if !w.connected || w.previousSSID == "" {
		return nil
	}

	fmt.Printf("Reconnecting to %q...\n", w.previousSSID)

	out, err := exec.Command("nmcli", "connection", "up", w.previousSSID).CombinedOutput()
	if err != nil {
		return fmt.Errorf("reconnecting to %s: %s", w.previousSSID, string(out))
	}

	w.connected = false
	time.Sleep(2 * time.Second)
	return nil
}

// DefaultTimeout is the max time to spend on Vespera WiFi before force-reconnecting.
const DefaultTimeout = 5 * time.Minute

// WithVespera connects to Vespera, runs the function with a timeout, then reconnects.
// Catches Ctrl+C, timeout, and errors — always reconnects.
func WithVespera(fn func() error) error {
	return WithVesperaTimeout(fn, DefaultTimeout)
}

// WithVesperaTimeout is like WithVespera but with a custom timeout.
func WithVesperaTimeout(fn func() error, timeout time.Duration) error {
	wm := NewWiFiManager()

	if err := wm.Connect(); err != nil {
		return err
	}

	// Catch Ctrl+C to ensure reconnect
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	var fnErr error
	select {
	case fnErr = <-done:
	case <-time.After(timeout):
		fnErr = fmt.Errorf("timed out after %s on Vespera WiFi", timeout)
	case sig := <-sigCh:
		fnErr = fmt.Errorf("interrupted by %s", sig)
	}

	if err := wm.Disconnect(); err != nil {
		fmt.Printf("Warning: failed to reconnect: %v\n", err)
	}

	return fnErr
}
