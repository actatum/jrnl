package tui

import (
	"github.com/actatum/jrnl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// SelectMsg the message to change the view to the selected entry.
type SelectMsg struct {
	EntryID int
}

type mode int

const (
	nav mode = iota
	edit
	create
)

// JournalUI implements tea.Model.
type JournalUI struct {
	mode      mode
	entryList list.Model
	input     textinput.Model
	quitting  bool
}

// InitJournalUI initializes the journalui model.
func InitJournalUI(jr *jrnl.Journal) (tea.Model, error) {
	input := textinput.New()
	input.Prompt = "> "
	input.Placeholder = "..."
	input.Width = 50

	items, err := newEntryList(jr)
	if err != nil {
		return nil, err
	}

	ui := JournalUI{
		mode:      nav,
		entryList: list.New(items, list.NewDefaultDelegate(), 0, 0),
		input:     input,
	}

	ui.entryList.Title = "Journal Entries"
	ui.entryList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keymap.Create,
			Keymap.Delete,
			Keymap.Back,
		}
	}

	return ui, nil
}

func (ui JournalUI) Init() tea.Cmd {
	return nil
}

func (ui JournalUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if ui.input.Focused() {

		} else {
			switch {
			case key.Matches(msg, Keymap.Quit):
				ui.quitting = true
				return ui, tea.Quit
			}
		}
	}

	return ui, tea.Batch(cmds...)
}

// View returns the text UI to be output to the terminal.
func (ui JournalUI) View() string {
	if ui.quitting {
		return ""
	}
	if ui.input.Focused() {
		return DocStyle.Render(ui.entryList.View() + "\n" + ui.input.View())
	}

	return DocStyle.Render(ui.entryList.View() + "\n")
}

func newEntryList(jr *jrnl.Journal) ([]list.Item, error) {
	entries, err := jr.ListEntries()
	if err != nil {
		return nil, err
	}

	return entriesToItems(entries), nil
}

func entriesToItems(entries []jrnl.Entry) []list.Item {
	items := make([]list.Item, 0, len(entries))
	//for _, entry := range entries {
	// items = append(items, list.Item(entry))
	//}

	return items
}
