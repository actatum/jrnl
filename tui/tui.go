package tui

import (
	"fmt"
	"os"

	"github.com/actatum/jrnl"
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the tui program.
func Run() error {
	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(basePath, os.ModePerm); err != nil {
		return err
	}

	dbPath := basePath + "/db"
	jr, err := jrnl.NewJournal(dbPath)
	if err != nil {
		return err
	}

	initialized, err := jr.IsInitialized()
	if err != nil {
		return err
	}

	if initialized {
		// enter password prompt
		pw := EnterPasswordPrompt()
		err = jr.Auth(pw)
		if err != nil {
			return err
		}
	} else {
		// create password prompt
		pw, err := CreatePasswordPrompt()
		if err != nil {
			return err
		}

		fmt.Println("pw = ", pw)
		err = jr.CreatePassword(pw)
		if err != nil {
			return err
		}
	}

	m, err := InitJournalUI(jr)
	if err != nil {
		return err
	}

	if _, err = tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		return err
	}

	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
