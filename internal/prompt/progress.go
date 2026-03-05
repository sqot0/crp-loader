package prompt

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/sqot0/crp-loader/internal"
)

type InstallTask struct {
	ID     string
	Name   string
	Status internal.InstallStatus
	Err    error
}

type progressModel struct {
	tasks    []*InstallTask
	spinner  spinner.Model
	quitting bool
}

func (m progressModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case internal.ProgressUpdate:
		for _, task := range m.tasks {
			if task.ID == msg.ID {
				task.Status = msg.Status
				task.Err = msg.Error
				break
			}
		}

		allFinished := true
		for _, task := range m.tasks {
			if task.Status != internal.StatusFinished && task.Status != internal.StatusError {
				allFinished = false
				break
			}
		}
		if allFinished {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m progressModel) View() tea.View {
	var sb strings.Builder
	sb.WriteString("\n")

	for _, task := range m.tasks {
		var icon string
		var statusText string

		switch task.Status {
		case internal.StatusWaiting:
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("○")
			statusText = "Ожидание..."
		case internal.StatusDownloading:
			icon = m.spinner.View()
			statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Загрузка...")
		case internal.StatusExtracting:
			icon = m.spinner.View()
			statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("190")).Render("Распаковка...")
		case internal.StatusFinished:
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓")
			statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("Готово")
		case internal.StatusError:
			icon = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("✗")
			statusText = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Ошибка: %v", task.Err))
		}

		sb.WriteString(fmt.Sprintf(" %s %-30s %s\n", icon, task.Name, statusText))
	}

	if m.quitting {
		sb.WriteString("\n Установка завершена!\n")
	}

	return tea.NewView(sb.String())
}

func RunProgress(tasks []*InstallTask, progressChan <-chan internal.ProgressUpdate) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	m := progressModel{
		tasks:   tasks,
		spinner: s,
	}

	p := tea.NewProgram(m)

	go func() {
		for update := range progressChan {
			p.Send(update)
		}
	}()

	_, err := p.Run()
	return err
}
