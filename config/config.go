package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(ConfigDir())

	viper.SetEnvPrefix("VESPERA")
	viper.AutomaticEnv()

	viper.SetDefault("host", "10.0.0.1")
	viper.SetDefault("ftp_port", 21)
	// Use XDG Pictures directory
	picturesDir := os.Getenv("XDG_PICTURES_DIR")
	if picturesDir == "" {
		out, err := exec.Command("xdg-user-dir", "PICTURES").Output()
		if err == nil {
			picturesDir = strings.TrimSpace(string(out))
		}
	}
	if picturesDir == "" {
		picturesDir = filepath.Join(homeDir(), "Images")
	}
	viper.SetDefault("output_dir", filepath.Join(picturesDir, "vespera"))

	_ = viper.ReadInConfig()
}

func ConfigDir() string {
	return filepath.Join(homeDir(), ".config", "vespera-cli")
}

func Host() string {
	return viper.GetString("host")
}

func FTPPort() int {
	return viper.GetInt("ftp_port")
}

func OutputDir() string {
	return viper.GetString("output_dir")
}

func Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return viper.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
