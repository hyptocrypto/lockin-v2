package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = ViewDetail
			m.err = nil
			return m, nil

		case "tab", "down":
			m.editInputs[m.editFocused].Blur()
			m.editFocused = (m.editFocused + 1) % len(m.editInputs)
			m.editInputs[m.editFocused].Focus()
			return m, textinput.Blink

		case "shift+tab", "backtab", "up":
			m.editInputs[m.editFocused].Blur()
			m.editFocused--
			if m.editFocused < 0 {
				m.editFocused = len(m.editInputs) - 1
			}
			m.editInputs[m.editFocused].Focus()
			return m, textinput.Blink

		case "enter":
			// Validate required fields
			name := m.editInputs[0].Value()
			username := m.editInputs[1].Value()
			password := m.editInputs[2].Value()

			if name == "" {
				m.err = fmt.Errorf("name is required")
				return m, nil
			}
			if password == "" {
				m.err = fmt.Errorf("password is required")
				return m, nil
			}

			// Update entry in vault
			entry := m.selected.ToStoreEntry()
			entry.Name = name
			entry.Username = username
			entry.Password = password
			entry.URL = m.editInputs[3].Value()
			entry.Notes = m.editInputs[4].Value()

			syncResult, err := m.Vault.Update(entry)
			if err != nil {
				m.err = fmt.Errorf("failed to update: %v", err)
				return m, nil
			}

			// Refresh passwords from vault
			_ = m.refreshPasswords()

			// Update selected entry
			updated := FromStoreEntry(entry)
			m.selected = &updated

			m.err = nil
			m.view = ViewDetail
			return m, m.setToast(formatSyncToast("Updated", name, syncResult))
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.editInputs[m.editFocused], cmd = m.editInputs[m.editFocused].Update(msg)
	return m, cmd
}

func (m Model) viewEdit() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render("✎ Edit Password")

	b.WriteString(header)
	b.WriteString("\n\n")

	// Input fields
	labels := []string{"Name *", "Username", "Password *", "URL", "Notes"}
	for i, input := range m.editInputs {
		style := blurredStyle
		if i == m.editFocused {
			style = focusedStyle
		}
		b.WriteString(style.Render(labels[i]))
		b.WriteString("\n")
		b.WriteString(input.View())
		b.WriteString("\n\n")
	}

	// Error message
	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())))
		b.WriteString("\n")
	}

	// Help
	b.WriteString(helpStyle.Render("Tab/↓ next • Shift+Tab/↑ prev • Enter save • Esc cancel"))

	// Center the content
	content := boxStyle.Width(50).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
