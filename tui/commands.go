package tui

import (
	"github.com/actatum/jrnl"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ error }
type updateEntryListMsg struct{}

func deleteEntryCmd(id int, jr *jrnl.Journal) tea.Cmd {
	return func() tea.Msg {
		err := jr.DeleteEntry(id)
		if err != nil {
			return errMsg{err}
		}
		return updateEntryListMsg{}
	}
}
