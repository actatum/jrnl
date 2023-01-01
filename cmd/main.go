package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/actatum/jrnl"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle           = lipgloss.NewStyle().Margin(1, 2)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

const layoutUS = "2006-01-02"

type journal struct {
	entries list.Model
}

func (j journal) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (j journal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return j, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		j.entries.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	j.entries, cmd = j.entries.Update(msg)
	return j, cmd
}

func (j journal) View() string {
	return docStyle.Render(j.entries.View())
}

type entry struct {
	title      string
	desc       string
	Content    string
	CreateTime time.Time
}

func initialEntry() entry {
	return entry{
		CreateTime: time.Now(),
	}
}

func (e entry) Title() string       { return e.title }
func (e entry) Description() string { return e.desc }
func (e entry) FilterValue() string { return e.Content }

func (e entry) Init() tea.Cmd {
	return nil
}

func (e entry) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return e, tea.Quit
		case "ctrl+s":
			return e, tea.Quit
		case "backspace":
			if len(e.Content) > 0 {
				e.Content = e.Content[:len(e.Content)-1]
			}
		default:
			e.Content += msg.String()
			// fmt.Println(e.Content)
		}
	}

	return e, nil
}

func (e entry) View() string {
	// The header
	s := "What's on your mind?\n\n"

	s += e.Content

	s += "\n\nPress ctrl+s to save, or ctrl+c to quit without saving.\n"

	return s
}

func newItemDelegate(keys *delegateKeyMap) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		var title string

		i, ok := m.SelectedItem().(entry)
		if ok {
			title = i.Title()
		} else {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keys.choose):
				fmt.Println("HELLO WORLD")
				i.View()
				_, cmd := i.Update(msg)
				return cmd
				//return m.NewStatusMessage(statusMessageStyle("You chose " + title))

			case key.Matches(msg, keys.remove):
				index := m.Index()
				m.RemoveItem(index)
				if len(m.Items()) == 0 {
					keys.remove.SetEnabled(false)
				}
				return m.NewStatusMessage(statusMessageStyle("Deleted " + title))
			}
		}

		return nil
	}

	help := []key.Binding{keys.choose, keys.remove}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

type delegateKeyMap struct {
	choose key.Binding
	remove key.Binding
}

func newDelegateKeyMap() *delegateKeyMap {
	return &delegateKeyMap{
		choose: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "choose"),
		),
		remove: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
	}
}

func (d delegateKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		d.choose,
		d.remove,
	}
}

func (d delegateKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			d.choose,
			d.remove,
		},
	}
}

func main() {
	f, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	jl, err := jrnl.NewJournal(f.Name())
	if err != nil {
		log.Fatal(err)
	}

	_, err = jl.CreateEntry("heres some content")
	if err != nil {
		log.Fatal(err)
	}

	_, err = jl.CreateEntry("second post")
	if err != nil {
		log.Fatal(err)
	}

	entries, err := jl.ListEntries()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(entries)

	delegate := newItemDelegate(newDelegateKeyMap())
	j := journal{entries: list.New([]list.Item{
		entry{
			title:      time.Now().Format(layoutUS),
			desc:       "abc",
			Content:    "some content",
			CreateTime: time.Time{},
		},
		entry{
			title:      time.Now().Add(-24 * time.Hour).Format(layoutUS),
			desc:       "123",
			Content:    "howdy",
			CreateTime: time.Time{},
		},
	}, delegate, 0, 0)}
	j.entries.Title = "Journal Entries"

	p := tea.NewProgram(j)
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
