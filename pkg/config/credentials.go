package config

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the identifier used for storing credentials in the system keyring
	ServiceName = "ancestrydl"
	// UsernameKey is the key for storing the username
	UsernameKey = "username"
	// PasswordKey is the key for storing the password
	PasswordKey = "password"
	// CookiesKey is the key for storing session cookies
	CookiesKey = "cookies"
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

// DeleteCredentials removes the stored credentials from the system keyring
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

	// Delete cookies
	if err := keyring.Delete(ServiceName, CookiesKey); err != nil {
		if err != keyring.ErrNotFound {
			errs = append(errs, fmt.Errorf("failed to delete cookies: %w", err))
		}
	}

	// If there were any errors, return them
	if len(errs) > 0 {
		return fmt.Errorf("error(s) deleting credentials: %v", errs)
	}

	return nil
}

// SaveCookies stores session cookies in the system keyring
func SaveCookies(cookiesJSON string) error {
	if cookiesJSON == "" {
		return fmt.Errorf("cookies data is empty")
	}

	if err := keyring.Set(ServiceName, CookiesKey, cookiesJSON); err != nil {
		return fmt.Errorf("failed to save cookies: %w", err)
	}

	return nil
}

// GetCookies retrieves stored session cookies from the system keyring
func GetCookies() (string, error) {
	cookies, err := keyring.Get(ServiceName, CookiesKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("no cookies found in keyring")
		}
		return "", fmt.Errorf("failed to retrieve cookies: %w", err)
	}

	return cookies, nil
}
