package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Confirm delete
			if m.deleteTarget != nil {
				name := m.deleteTarget.Name
				syncResult, err := m.Vault.Delete(m.deleteTarget.ID)
				if err == nil {
					_ = m.refreshPasswords()
					m.selected = nil
					m.deleteTarget = nil
					// Adjust cursor if needed
					if m.cursor >= len(m.passwords) && m.cursor > 0 {
						m.cursor--
					}
					m.view = ViewList
					return m, m.setToast(formatSyncToast("Deleted", name, syncResult))
				}
			}
			m.view = ViewList
			return m, nil

		case "n", "N", "esc", "q":
			// Cancel delete - go back to detail if we have a selected entry, otherwise list
			m.deleteTarget = nil
			if m.selected != nil {
				m.view = ViewDetail
			} else {
				m.view = ViewList
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) viewConfirmDelete() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(errorColor).
		MarginBottom(1).
		Render("⚠ Confirm Delete")

	b.WriteString(header)
	b.WriteString("\n\n")

	if m.deleteTarget != nil {
		// Warning message
		msg := fmt.Sprintf("Are you sure you want to delete '%s'?", m.deleteTarget.Name)
		b.WriteString(lipgloss.NewStyle().Foreground(textColor).Render(msg))
		b.WriteString("\n\n")

		b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render("This action cannot be undone."))
		b.WriteString("\n\n")
	}

	// Options
	yesStyle := lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	noStyle := lipgloss.NewStyle().
		Foreground(secondaryColor).
		Bold(true)

	b.WriteString(yesStyle.Render("[Y]"))
	b.WriteString(lipgloss.NewStyle().Foreground(textColor).Render(" Yes, delete"))
	b.WriteString("    ")
	b.WriteString(noStyle.Render("[N]"))
	b.WriteString(lipgloss.NewStyle().Foreground(textColor).Render(" No, cancel"))

	// Help
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Y confirm • N/Esc cancel"))

	// Center the content
	content := boxStyle.Width(50).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
