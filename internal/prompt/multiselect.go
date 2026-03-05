package prompt

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type multiItem struct {
	text    string
	checked bool
}

type multiModel struct {
	items    []multiItem
	cursor   int
	quitting bool
	styles   multiStyles
}

type multiStyles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	desc         lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

func newMultiStyles() multiStyles {
	var s multiStyles
	s.title = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.desc = lipgloss.NewStyle().PaddingLeft(6).Foreground(lipgloss.BrightBlack)
	s.pagination = lipgloss.NewStyle().PaddingLeft(4) // simplified
	s.help = lipgloss.NewStyle().PaddingLeft(4).PaddingBottom(1).Foreground(lipgloss.Color("241"))
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

func initialMultiModel(choices []string) multiModel {
	items := make([]multiItem, len(choices))
	for i, c := range choices {
		items[i] = multiItem{text: c, checked: false}
	}
	m := multiModel{items: items, cursor: 0, quitting: false}
	m.styles = newMultiStyles()
	return m
}

func (m multiModel) Init() tea.Cmd {
	return nil
}

func (m multiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "space":
			m.items[m.cursor].checked = !m.items[m.cursor].checked
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiModel) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}
	var b strings.Builder
	b.WriteString("\n" + m.styles.title.Render("Выберите опциональные наборы:") + "\n\n")
	for i, item := range m.items {
		checked := " "
		if item.checked {
			checked = "✓"
		}
		lines := strings.Split(item.text, "\n")
		name := lines[0]
		desc := ""
		if len(lines) > 1 {
			desc = lines[1]
		}
		if i == m.cursor {
			b.WriteString(m.styles.selectedItem.Render(fmt.Sprintf("  > [%s] %s", checked, name)) + "\n")
			if desc != "" {
				b.WriteString(m.styles.desc.Render(desc) + "\n")
			}
		} else {
			b.WriteString(m.styles.item.Render(fmt.Sprintf("  [%s] %s", checked, name)) + "\n")
			if desc != "" {
				b.WriteString(m.styles.desc.Render(desc) + "\n")
			}
		}
	}
	b.WriteString("\n" + m.styles.help.Render("↑ ↓ — навигация • Пробел — Выбрать • Enter — Продолжить"))
	return tea.NewView(b.String())
}

func Multiselect(choices []string) ([]int, error) {
	m := initialMultiModel(choices)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}
	fm, ok := finalModel.(multiModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}
	var selected []int
	for i, item := range fm.items {
		if item.checked {
			selected = append(selected, i)
		}
	}
	return selected, nil
}
