package prompt

import (
	"fmt"
	"io"
	"strings"

	. "charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const listHeight = 7

type styles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newStyles(darkBG bool) styles {
	var s styles
	s.title = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.pagination = DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1).Foreground(lipgloss.Color("241"))
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type itemDelegate struct {
	styles *styles
}

func (d itemDelegate) Height() int                          { return 1 }
func (d itemDelegate) Spacing() int                         { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m Model, index int, item Item) {
	i := item

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list     Model
	choice   string
	styles   styles
	quitting bool
}

func initialModel(choices []string, title string) model {
	items := make([]Item, len(choices))
	for i, c := range choices {
		items[i] = item(c)
	}

	const defaultWidth = 20

	l := New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	m := model{list: l}
	m.updateStyles(true) // default to dark styles.
	return m
}

func (m *model) updateStyles(isDark bool) {
	m.styles = newStyles(isDark)
	m.list.Styles.Title = m.styles.title
	m.list.Styles.PaginationStyle = m.styles.pagination
	m.list.Styles.HelpStyle = m.styles.help
	m.list.SetDelegate(itemDelegate{styles: &m.styles})
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyPressMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	return tea.NewView("\n" + m.list.View() + "\n" + m.styles.help.Render("↑ ↓ — навигация • Enter — выбрать"))
}

func Select(choices []string, title string) (string, error) {
	finalModel, err := tea.NewProgram(initialModel(choices, title)).Run()
	if err != nil {
		return "", err
	}
	fm, ok := finalModel.(model)
	if !ok {
		return "", fmt.Errorf("unexpected model type")
	}
	return fm.choice, nil
}

type item string

func (i item) FilterValue() string { return "" }
func (i item) String() string      { return string(i) }
