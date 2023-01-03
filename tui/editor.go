package tui

import (
	"fmt"
	"log"

	"github.com/actatum/jrnl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EditorUI implements tea.Model.
type EditorUI struct {
	entry        entryItem
	updatedEntry entryItem
	textarea     textarea.Model
	jr           *jrnl.Journal
	create       bool
	quitting     bool
}

// InitEditorUI ...
func InitEditorUI(e entryItem, jr *jrnl.Journal, create bool) tea.Model {
	ui := EditorUI{
		entry:        e,
		updatedEntry: e,
		textarea:     textarea.New(),
		jr:           jr,
		create:       create,
	}

	ui.textarea.SetValue(e.Content)
	ui.textarea.CharLimit = 50000
	ui.textarea.Focus()
	ui.textarea.SetWidth(WindowSize.Width)
	ui.textarea.SetHeight(WindowSize.Height - ui.verticalMarginHeight())

	return ui
}

// Init ...
func (ui EditorUI) Init() tea.Cmd {
	return textarea.Blink
}

// Update ...
func (ui EditorUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case createEntryMsg:
		m, err := InitEntryUI(msg.entry, ui.jr)
		if err != nil {
			return ui, func() tea.Msg { return errMsg{err} }
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Back):
			m, err := InitEntryUI(ui.entry, ui.jr)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			return m, tea.Batch(cmds...)
		case key.Matches(msg, Keymap.ForceQuit):
			return ui, tea.Quit
		case key.Matches(msg, Keymap.Save):
			if ui.create {
				cmds = append(cmds, createEntryCmd(ui.updatedEntry.Content, ui.jr))
			} else {
				cmds = append(cmds, editEntryCmd(ui.updatedEntry, ui.jr))
			}
			m, err := InitEntryUI(ui.updatedEntry, ui.jr)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			return m, tea.Batch(cmds...)
		default:
			ui.updatedEntry.Content = ui.textarea.Value()
			ui.textarea, cmd = ui.textarea.Update(msg)
			cmds = append(cmds, cmd)
			if !ui.textarea.Focused() {
				cmd = ui.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	case tea.WindowSizeMsg:
		WindowSize = msg
		ui.textarea.SetWidth(msg.Width)
		ui.textarea.SetHeight(msg.Height - ui.verticalMarginHeight())
	case errMsg:
		log.Println(msg.Error())
	}

	//ui.textarea, cmd = ui.textarea.Update(msg)
	//ui.updatedEntry.Content = ui.textarea.Value()
	//cmds = append(cmds, cmd)
	return ui, tea.Batch(cmds...)
}

// View ...
func (ui EditorUI) View() string {
	if ui.quitting {
		return ""
	}

	return fmt.Sprintf("%s\n%s", ui.textarea.View(), ui.helpView())
}

func (ui EditorUI) helpView() string {
	// TODO: use the keymaps to populate the help string
	return HelpStyle("\n • ctrl+s save • esc back \n")
}

func (ui EditorUI) verticalMarginHeight() int {
	helpHeight := lipgloss.Height(ui.helpView())
	return helpHeight
}
