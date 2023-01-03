package tui

import (
	"bytes"
	"fmt"
	"syscall"

	"golang.org/x/term"
)

// CreatePasswordPrompt prompts the user to create a new password for their journal.
func CreatePasswordPrompt() (string, error) {
	fmt.Println("Create a password for your journal...")
	pw, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}
	fmt.Println("Re-enter your password...")
	reentry, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	if !bytes.Equal(pw, reentry) {
		return "", fmt.Errorf("passwords don't match")
	}

	return string(pw), nil
}

// EnterPasswordPrompt prompts the user to enter the password for their journal.
func EnterPasswordPrompt() (string, error) {
	fmt.Println("Enter your journal password...")
	pw, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return "", err
	}

	return string(pw), nil
}
