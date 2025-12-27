package ui

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.selected = nil
			m.toastText = ""
			m.view = ViewList
			return m, nil

		case "c":
			if m.selected != nil {
				if err := clipboard.WriteAll(m.selected.Password); err == nil {
					return m, m.setToast("✓ Password copied to clipboard")
				}
				return m, m.setToast("✗ Failed to copy password")
			}
			return m, nil

		case "u":
			if m.selected != nil {
				if err := clipboard.WriteAll(m.selected.Username); err == nil {
					return m, m.setToast("✓ Username copied to clipboard")
				}
				return m, m.setToast("✗ Failed to copy username")
			}
			return m, nil

		case "e":
			if m.selected != nil {
				// Populate edit inputs with current values
				m.editInputs[0].SetValue(m.selected.Name)
				m.editInputs[1].SetValue(m.selected.Username)
				m.editInputs[2].SetValue(m.selected.Password)
				m.editInputs[3].SetValue(m.selected.URL)
				m.editInputs[4].SetValue(m.selected.Notes)

				// Focus the first input
				m.editFocused = 0
				for i := range m.editInputs {
					m.editInputs[i].Blur()
				}
				m.editInputs[0].Focus()

				m.view = ViewEdit
			}
			return m, nil

		case "d":
			if m.selected != nil {
				m.deleteTarget = m.selected
				m.view = ViewConfirmDelete
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) viewDetail() string {
	var b strings.Builder

	if m.selected == nil {
		return "No password selected"
	}

	entry := m.selected

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render(fmt.Sprintf("🔑 %s", entry.Name))

	b.WriteString(header)
	b.WriteString("\n\n")

	// Details
	fieldStyle := lipgloss.NewStyle().
		Foreground(mutedColor).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(textColor)

	// Username
	if entry.Username != "" {
		b.WriteString(fieldStyle.Render("Username:"))
		b.WriteString(valueStyle.Render(entry.Username))
		b.WriteString("\n")
	}

	// Password (masked)
	b.WriteString(fieldStyle.Render("Password:"))
	masked := strings.Repeat("•", len(entry.Password))
	b.WriteString(valueStyle.Render(masked))
	b.WriteString("\n")

	// URL
	if entry.URL != "" {
		b.WriteString(fieldStyle.Render("URL:"))
		b.WriteString(valueStyle.Render(entry.URL))
		b.WriteString("\n")
	}

	// Notes
	if entry.Notes != "" {
		b.WriteString("\n")
		b.WriteString(fieldStyle.Render("Notes:"))
		b.WriteString("\n")
		b.WriteString(valueStyle.Render(entry.Notes))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("c copy password • u copy username • e edit • d delete • Esc/q back"))

	// Center the content
	content := boxStyle.Width(50).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
