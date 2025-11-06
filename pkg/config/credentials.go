package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the identifier used for storing credentials in the system keyring
	ServiceName = "ancestrydl"
	// UsernameKey is the key for storing the username
	UsernameKey = "username"
	// PasswordKey is the key for storing the password
	PasswordKey = "password"
	// ConfigDirName is the name of the config directory
	ConfigDirName = ".ancestrydl"
	// CookiesFileName is the name of the cookies file
	CookiesFileName = "cookies.json"
	// ConfigFileName is the name of the config file
	ConfigFileName = "config.json"
)

var (
	// ErrCredentialsNotFound is returned when no credentials are stored
	ErrCredentialsNotFound = errors.New("credentials not found in keyring")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials provided")
)

// Credentials represents a user's Ancestry.com login information
type Credentials struct {
	Username string
	Password string
}

// SaveCredentials stores the username and password in the system keyring
func SaveCredentials(username, password string) error {
	if username == "" || password == "" {
		return ErrInvalidCredentials
	}

	// Store username
	if err := keyring.Set(ServiceName, UsernameKey, username); err != nil {
		return fmt.Errorf("failed to save username: %w", err)
	}

	// Store password
	if err := keyring.Set(ServiceName, PasswordKey, password); err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	return nil
}

// GetCredentials retrieves the stored username and password from the system keyring
func GetCredentials() (*Credentials, error) {
	// Retrieve username
	username, err := keyring.Get(ServiceName, UsernameKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrCredentialsNotFound
		}
		return nil, fmt.Errorf("failed to retrieve username: %w", err)
	}

	// Retrieve password
	password, err := keyring.Get(ServiceName, PasswordKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, ErrCredentialsNotFound
		}
		return nil, fmt.Errorf("failed to retrieve password: %w", err)
	}

	return &Credentials{
		Username: username,
		Password: password,
	}, nil
}

// DeleteCredentials removes the stored credentials from the system keyring and cookies file
func DeleteCredentials() error {
	var errs []error

	// Delete username
	if err := keyring.Delete(ServiceName, UsernameKey); err != nil {
		if err != keyring.ErrNotFound {
			errs = append(errs, fmt.Errorf("failed to delete username: %w", err))
		}
	}

	// Delete password
	if err := keyring.Delete(ServiceName, PasswordKey); err != nil {
		if err != keyring.ErrNotFound {
			errs = append(errs, fmt.Errorf("failed to delete password: %w", err))
		}
	}

	// Delete cookies file
	cookiesPath, err := getCookiesFilePath()
	if err == nil {
		if err := os.Remove(cookiesPath); err != nil {
			if !os.IsNotExist(err) {
				errs = append(errs, fmt.Errorf("failed to delete cookies file: %w", err))
			}
		}
	}

	// If there were any errors, return them
	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting credentials: %v", errs)
	}

	return nil
}

// getConfigDir returns the path to the config directory (~/.ancestrydl)
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDirName)
	return configDir, nil
}

// getCookiesFilePath returns the full path to the cookies file
func getCookiesFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, CookiesFileName), nil
}

// ensureConfigDir creates the config directory if it doesn't exist
func ensureConfigDir() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create directory with secure permissions (0700 - only owner can access)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// SaveCookies stores session cookies in ~/.ancestrydl/cookies.json
func SaveCookies(cookiesJSON string) error {
	if cookiesJSON == "" {
		return fmt.Errorf("cookies data is empty")
	}

	// Ensure config directory exists
	if err := ensureConfigDir(); err != nil {
		return err
	}

	// Get cookies file path
	cookiesPath, err := getCookiesFilePath()
	if err != nil {
		return err
	}

	// Write cookies to file with secure permissions (0600 - only owner can read/write)
	if err := os.WriteFile(cookiesPath, []byte(cookiesJSON), 0600); err != nil {
		return fmt.Errorf("failed to write cookies file: %w", err)
	}

	fmt.Printf("   Cookies saved to: %s (%d bytes)\n", cookiesPath, len(cookiesJSON))

	return nil
}

// GetCookies retrieves stored session cookies from ~/.ancestrydl/cookies.json
func GetCookies() (string, error) {
	cookiesPath, err := getCookiesFilePath()
	if err != nil {
		return "", err
	}

	// Read cookies from file
	data, err := os.ReadFile(cookiesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no cookies found - please run 'ancestrydl login' first")
		}
		return "", fmt.Errorf("failed to read cookies file: %w", err)
	}

	return string(data), nil
}

// Config represents user configuration settings
type Config struct {
	DefaultTreeID string `json:"defaultTreeId,omitempty"`
}

// getConfigFilePath returns the full path to the config file
func getConfigFilePath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, ConfigFileName), nil
}

// SaveConfig stores configuration settings in ~/.ancestrydl/config.json
func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Ensure config directory exists
	if err := ensureConfigDir(); err != nil {
		return err
	}

	// Get config file path
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config to file with secure permissions
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig retrieves configuration settings from ~/.ancestrydl/config.json
func GetConfig() (*Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	// Read config from file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// SetDefaultTreeID sets the default tree ID in the config
func SetDefaultTreeID(treeID string) error {
	cfg, err := GetConfig()
	if err != nil {
		return err
	}

	cfg.DefaultTreeID = treeID
	return SaveConfig(cfg)
}

// GetDefaultTreeID retrieves the default tree ID from config
func GetDefaultTreeID() (string, error) {
	cfg, err := GetConfig()
	if err != nil {
		return "", err
	}

	return cfg.DefaultTreeID, nil
}
