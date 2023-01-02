package tui

import (
	"bufio"
	"fmt"
	"os"
)

// CreatePasswordPrompt prompts the user to create a new password for their journal.
func CreatePasswordPrompt() (string, error) {
	var pw string
	fmt.Println("Create a password for your journal")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	pw = scanner.Text()
	fmt.Println("Re-enter your password")
	var reentry string
	scanner.Scan()
	reentry = scanner.Text()

	if pw != reentry {
		return "", fmt.Errorf("passwords don't match")
	}

	return pw, nil
}

// EnterPasswordPrompt prompts the user to enter the password for their journal.
func EnterPasswordPrompt() string {
	var pw string
	fmt.Println("Enter your journal password...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	pw = scanner.Text()

	return pw
}
