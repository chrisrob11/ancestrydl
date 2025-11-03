package ancestry

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	// LoginURL is the Ancestry.com login page
	LoginURL = "https://www.ancestry.com/account/signin"

	// TwoFactorMethodEmail represents email-based 2FA
	TwoFactorMethodEmail = "email"
	// TwoFactorMethodPhone represents phone/SMS-based 2FA
	TwoFactorMethodPhone = "phone"
)

// fillInputField clicks, clears, and fills an input field
func fillInputField(field *rod.Element, value, fieldName string) error {
	// Click to focus the field
	if err := field.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click %s field: %w", fieldName, err)
	}
	time.Sleep(500 * time.Millisecond)

	// Set the value using JavaScript and trigger proper events
	escapedValue := fmt.Sprintf("%q", value)
	script := fmt.Sprintf(`() => {
		const input = this;
		input.value = '';
		input.focus();

		// Set the value
		const value = %s;
		input.value = value;

		// Trigger input and change events so the form knows it changed
		input.dispatchEvent(new Event('input', { bubbles: true }));
		input.dispatchEvent(new Event('change', { bubbles: true }));

		return input.value.length;
	}`, escapedValue)

	_, err := field.Evaluate(rod.Eval(script))
	if err != nil {
		return fmt.Errorf("failed to set %s value: %w", fieldName, err)
	}

	time.Sleep(500 * time.Millisecond)
	return nil
}

// LoginOptions contains options for the Login function
type LoginOptions struct {
	SkipSubmit      bool   // If true, fills the form but doesn't submit
	TwoFactorMethod string // TwoFactorMethodEmail or TwoFactorMethodPhone - which 2FA method to automatically select
}

// Login authenticates to Ancestry.com with the provided credentials
func (c *Client) Login(username, password string) error {
	return c.LoginWithOptions(username, password, LoginOptions{})
}

// LoginWithOptions authenticates to Ancestry.com with the provided credentials and options
func (c *Client) LoginWithOptions(username, password string, opts LoginOptions) error {
	if c.page == nil {
		return fmt.Errorf("no page available, call NavigateToAncestry first")
	}

	// Navigate to login page
	if err := c.page.Navigate(LoginURL); err != nil {
		return fmt.Errorf("failed to navigate to login page: %w", err)
	}

	if err := c.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for login page load: %w", err)
	}

	// Wait for the login form to be visible
	time.Sleep(2 * time.Second)

	// Find and fill username field
	usernameField, err := c.page.Element("#username")
	if err != nil {
		return fmt.Errorf("failed to find username field: %w", err)
	}
	if err := fillInputField(usernameField, username, "username"); err != nil {
		return err
	}

	// Find and fill password field
	passwordField, err := c.page.Element("#password")
	if err != nil {
		return fmt.Errorf("failed to find password field: %w", err)
	}
	if err := fillInputField(passwordField, password, "password"); err != nil {
		return err
	}

	// If skip submit is enabled, stop here
	if opts.SkipSubmit {
		return nil
	}

	return c.submitLoginAndVerify(opts)
}

// submitLoginAndVerify submits the login form and verifies success
func (c *Client) submitLoginAndVerify(opts LoginOptions) error {
	// Find and click the sign in button
	signInButton, err := c.page.Element("button[type='submit']")
	if err != nil {
		return fmt.Errorf("failed to find sign in button: %w", err)
	}

	if err := signInButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click sign in button: %w", err)
	}

	// Wait for navigation after login
	time.Sleep(2 * time.Second)

	// Wait for URL to change or 2FA buttons to appear (up to 10 seconds)
	maxWaitTime := 10 * time.Second
	checkInterval := 500 * time.Millisecond
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		currentURL := c.page.MustInfo().URL

		// Check if we've navigated away from the initial login page
		if currentURL != LoginURL {
			break
		}

		// Or check if 2FA buttons appeared (sometimes URL stays the same)
		buttons, err := c.page.Elements("button.methodBtn")
		if err == nil && len(buttons) > 0 {
			break
		}

		time.Sleep(checkInterval)
	}

	// Final wait for page to fully load
	time.Sleep(1 * time.Second)

	// Check if login was successful by looking for error messages or success indicators
	currentURL := c.page.MustInfo().URL

	// FIRST: Check if we're on a 2FA selection page (this takes priority)
	// Look for 2FA method buttons to determine if 2FA is required
	allButtons, checkErr := c.page.Elements("button.methodBtn")
	is2FARequired := checkErr == nil && len(allButtons) > 0

	if is2FARequired {
		if opts.TwoFactorMethod == "" {
			// No 2FA method specified - fail and tell user to specify one
			return fmt.Errorf("two-factor authentication is required. Please specify a 2FA method using the --2fa flag with either '%s' or '%s'", TwoFactorMethodEmail, TwoFactorMethodPhone)
		}

		// 2FA method specified - handle it
		if err := c.handle2FA(opts.TwoFactorMethod); err != nil {
			return fmt.Errorf("2FA handling failed: %w", err)
		}

		return nil
	}

	// SECOND: If we're still on the login page and no 2FA, login likely failed
	if currentURL == LoginURL || c.page.MustHas("div.errorMessage") {
		// Try to get error message if available
		errorElem, err := c.page.Element("div.errorMessage")
		if err == nil {
			errorText, _ := errorElem.Text()
			return fmt.Errorf("login failed: %s", errorText)
		}
		return fmt.Errorf("login failed: invalid credentials or unexpected error")
	}

	// Login appears successful
	return nil
}

// handle2FA automatically selects the 2FA method (email or phone) if prompted
func (c *Client) handle2FA(method string) error {
	// Wait for 2FA page to fully load
	time.Sleep(3 * time.Second)

	// Determine which data-method to look for
	var dataMethodValue string
	switch method {
	case TwoFactorMethodEmail:
		dataMethodValue = TwoFactorMethodEmail
	case TwoFactorMethodPhone:
		dataMethodValue = "sms"
	default:
		return fmt.Errorf("invalid 2FA method: %s (must be '%s' or '%s')", method, TwoFactorMethodEmail, TwoFactorMethodPhone)
	}

	fmt.Printf("Selecting 2FA method: %s\n", method)

	// Use JavaScript to click the button
	result, err := c.page.Evaluate(rod.Eval(fmt.Sprintf(`
		() => {
			const button = document.querySelector("button[data-method='%s']");
			if (!button) {
				return { success: false, error: "Button not found" };
			}

			// Scroll to button
			button.scrollIntoView({ behavior: 'smooth', block: 'center' });

			// Try clicking parent container first (the whole row)
			const parent = button.closest('.ancCol, .methodOption, div[class*="method"]');
			if (parent) {
				parent.click();
			}

			// Dispatch mouse events and click
			button.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
			button.dispatchEvent(new MouseEvent('mouseup', { bubbles: true, cancelable: true }));
			button.click();

			return { success: true, error: null };
		}
	`, dataMethodValue)))

	if err != nil {
		return fmt.Errorf("failed to select 2FA method: %w", err)
	}

	// Check the result
	resultMap := result.Value.Map()
	if success, ok := resultMap["success"]; !ok || !success.Bool() {
		if errorMsg, ok := resultMap["error"]; ok && errorMsg.Str() != "" {
			return fmt.Errorf("2FA method selection failed: %s", errorMsg.Str())
		}
		return fmt.Errorf("2FA method selection failed")
	}

	// Wait for the 2FA code entry page to load
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== WAITING FOR 2FA CODE ===")
	fmt.Println("Please enter the verification code in the browser when you receive it.")
	fmt.Println("Polling every 3 seconds to detect when authentication completes...")

	// Wait for user to enter code and submit (or timeout after 3 minutes)
	maxAttempts := 60 // 60 * 3 seconds = 3 minutes
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(3 * time.Second)

		currentURL := c.page.MustInfo().URL

		// Check if we've moved away from the 2FA pages
		if !strings.Contains(currentURL, "/signin") && !strings.Contains(currentURL, "/mfa") {
			fmt.Printf("\n✓ Navigation detected! New URL: %s\n", currentURL)
			fmt.Println("2FA authentication completed successfully!")
			return nil
		}

		// Check if input fields are gone (alternative success indicator)
		hasInputs := c.page.MustHas("input[type='text']") || c.page.MustHas("input[type='tel']")
		if !hasInputs && !strings.Contains(currentURL, "/mfa") {
			fmt.Println("\n✓ 2FA page elements disappeared - authentication completed!")
			return nil
		}

		// Show progress every 15 seconds
		if (i+1)%5 == 0 {
			elapsed := (i + 1) * 3
			remaining := (maxAttempts - i - 1) * 3
			fmt.Printf("Still waiting... (%ds elapsed, %ds remaining)\n", elapsed, remaining)
		}
	}

	return fmt.Errorf("2FA timeout: user did not complete verification within 3 minutes")
}

// IsLoggedIn checks if the user is currently authenticated
func (c *Client) IsLoggedIn() bool {
	if c.page == nil {
		return false
	}

	// Check for elements that indicate logged-in state
	// This could be a user menu, account link, etc.
	_, err := c.page.Element("a[href*='/account']")
	return err == nil
}
