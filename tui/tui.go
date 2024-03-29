// Package tui provides types and functions for running the terminal ui that interacts with the journal.
package tui

import (
	"fmt"
	"log"
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

	var f *os.File
	if f, err = tea.LogToFile(basePath+"/debug.log", ""); err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	} else {
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	dbPath := basePath + "/db"
	jr, err := jrnl.NewJournal(dbPath)
	if err != nil {
		return err
	}
	defer func(jr *jrnl.Journal) {
		err = jr.Close()
		if err != nil {
			log.Printf("error closing journal: %v\n", err)
		}
	}(jr)

	initialized, err := jr.IsInitialized()
	if err != nil {
		return err
	}

	var pw string
	if initialized {
		// enter password prompt
		pw, err = EnterPasswordPrompt()
		if err != nil {
			return err
		}
	} else {
		// create password prompt
		pw, err = CreatePasswordPrompt()
		if err != nil {
			return err
		}

		err = jr.CreatePassword(pw)
		if err != nil {
			return err
		}
	}

	err = jr.Auth(pw)
	if err != nil {
		return err
	}

	m, err := InitJournalUI(jr)
	if err != nil {
		return err
	}

	if _, err = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		return err
	}

	return nil
}
