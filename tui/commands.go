package tui

import (
	"github.com/actatum/jrnl"
	tea "github.com/charmbracelet/bubbletea"
)

type errMsg struct{ error }
type updateEntryListMsg struct{}
type createEntryMsg struct {
	entry entryItem
}

func deleteEntryCmd(id int, jr *jrnl.Journal) tea.Cmd {
	return func() tea.Msg {
		err := jr.DeleteEntry(id)
		if err != nil {
			return errMsg{err}
		}
		return updateEntryListMsg{}
	}
}

func editEntryCmd(e entryItem, jr *jrnl.Journal) tea.Cmd {
	return func() tea.Msg {
		_, err := jr.EditEntry(e.ID, e.Content)
		if err != nil {
			return errMsg{err}
		}

		return nil
	}
}

func createEntryCmd(content string, jr *jrnl.Journal) tea.Cmd {
	return func() tea.Msg {
		entry, err := jr.CreateEntry(content)
		if err != nil {
			return errMsg{err}
		}

		return createEntryMsg{entryItem(entry)}
	}
}
