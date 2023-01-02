package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/actatum/jrnl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

// EntryUI implements tea.Model.
type EntryUI struct {
	entry    entryItem
	viewport viewport.Model
	jr       *jrnl.Journal
	renderer *glamour.TermRenderer
	ready    bool
	quitting bool
}

// InitEntryUI ...
func InitEntryUI(e entryItem, jr *jrnl.Journal) (tea.Model, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(WindowSize.Width-5),
		glamour.WithEmoji(),
	)
	if err != nil {
		return nil, err
	}

	ui := EntryUI{
		entry:    e,
		jr:       jr,
		renderer: renderer,
	}

	return ui, nil
}

// Init ...
func (ui EntryUI) Init() tea.Cmd {
	return nil
}

// Update ...
func (ui EntryUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keymap.Quit):
			return ui, tea.Quit
		case key.Matches(msg, Keymap.Back):
			m, err := InitJournalUI(ui.jr)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			return m, tea.Batch(cmds...)
		}
	case tea.WindowSizeMsg:
		WindowSize = msg
		headerHeight := lipgloss.Height(ui.headerView())
		footerHeight := lipgloss.Height(ui.footerView())
		helpHeight := lipgloss.Height(ui.helpView())
		verticalMarginHeight := headerHeight + footerHeight + helpHeight

		if !ui.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			var err error
			ui.renderer, err = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(msg.Width-5),
				glamour.WithEmoji(),
			)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			ui.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			ui.viewport.YPosition = headerHeight
			ui.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			str, err := ui.renderer.Render(ui.entry.Content)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			ui.viewport.SetContent(str)
			ui.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			ui.viewport.YPosition = headerHeight + 1
		} else {
			var err error
			ui.renderer, err = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(msg.Width-5),
				glamour.WithEmoji(),
			)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			str, err := ui.renderer.Render(ui.entry.Content)
			if err != nil {
				return ui, func() tea.Msg { return errMsg{err} }
			}
			ui.viewport.SetContent(str)
			ui.viewport.Width = msg.Width
			ui.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// Handle keyboard and mouse events in the viewport
	ui.viewport, cmd = ui.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return ui, tea.Batch(cmds...)
}

// View returns the text UI to be output to the terminal.
func (ui EntryUI) View() string {
	if ui.quitting {
		return ""
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", ui.headerView(), ui.viewport.View(), ui.footerView(), ui.helpView())
}

func (ui EntryUI) headerView() string {
	title := titleStyle.Render(ui.entry.CreateTime.Format(time.RFC822))
	line := strings.Repeat("─", max(0, ui.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (ui EntryUI) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", ui.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, ui.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (ui EntryUI) helpView() string {
	// TODO: use the keymaps to populate the help string
	return HelpStyle("\n ↑/↓: scroll • e: edit • esc: back • q: quit\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
