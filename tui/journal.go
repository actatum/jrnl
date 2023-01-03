package tui

import (
	"log"
	"strings"
	"time"

	"github.com/actatum/jrnl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const layoutUS = "2006-01-02"

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
	jr        *jrnl.Journal
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
		jr:        jr,
	}

	ui.entryList.Title = "Journal Entries"
	ui.entryList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			Keymap.Create,
			Keymap.Delete,
		}
	}
	top, right, bottom, left := DocStyle.GetMargin()
	ui.entryList.SetSize(WindowSize.Width-left-right, WindowSize.Height-top-bottom-1)

	return ui, nil
}

// Init ...
func (ui JournalUI) Init() tea.Cmd {
	return nil
}

// Update ...
func (ui JournalUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowSize = msg
		top, right, bottom, left := DocStyle.GetMargin()
		ui.entryList.SetSize(msg.Width-left-right, msg.Height-top-bottom-1)
	case updateEntryListMsg:
		entries, err := ui.jr.ListEntries()
		if err != nil {
			return ui, func() tea.Msg { return errMsg{err} }
		}
		items := entriesToItems(entries)
		ui.entryList.SetItems(items)
		ui.mode = nav
	case errMsg:
		log.Printf("ERROR: %s\n", msg.Error())
	case tea.KeyMsg:
		if ui.input.Focused() {
			switch {
			case key.Matches(msg, Keymap.Back):
				ui.input.SetValue("")
				ui.mode = nav
				ui.input.Blur()
			case key.Matches(msg, Keymap.Enter):
				if strings.ToLower(ui.input.Value()) == "delete" {
					cmds = append(cmds, deleteEntryCmd(ui.getActiveEntryID(), ui.jr))
					ui.input.SetValue("")
					ui.mode = nav
					ui.input.Blur()
				} else {
					ui.input.SetValue("")
					ui.mode = nav
					ui.input.Blur()
				}
			}
			ui.input, cmd = ui.input.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			switch {
			case key.Matches(msg, Keymap.Quit):
				ui.quitting = true
				return ui, tea.Quit
			case key.Matches(msg, Keymap.Create):
				return InitEditorUI(entryItem{}, ui.jr, true), tea.Batch(cmds...)
			case key.Matches(msg, Keymap.Enter):
				activeEntry := ui.entryList.SelectedItem().(entryItem)
				entry, err := InitEntryUI(activeEntry, ui.jr)
				if err != nil {
					return ui, func() tea.Msg { return errMsg{err} }
				}

				return entry.Update(WindowSize)
			case key.Matches(msg, Keymap.Delete):
				items := ui.entryList.Items()
				if len(items) > 0 {
					ui.input.Placeholder = "Type 'delete' to delete this entry\n"
					ui.input.Focus()
				}
			default:
				ui.entryList, cmd = ui.entryList.Update(msg)
			}
			cmds = append(cmds, cmd)
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

func (ui JournalUI) getActiveEntryID() int {
	items := ui.entryList.Items()
	activeItem := items[ui.entryList.Index()]
	return activeItem.(entryItem).ID
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
	for _, entry := range entries {
		items = append(items, list.Item(entryItem(entry)))
	}

	return items
}

type entryItem struct {
	ID         int
	Content    string
	CreateTime time.Time
	UpdateTime time.Time
}

func (i entryItem) Title() string       { return i.CreateTime.Format(time.RFC822) }
func (i entryItem) Description() string { return i.Content }
func (i entryItem) FilterValue() string { return i.CreateTime.Format(time.RFC822) }
