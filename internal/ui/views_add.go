package ui

import (
	"fmt"
	"lockin/internal/store"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = ViewList
			m.err = nil
			return m, nil

		case "tab", "down":
			m.addInputs[m.addFocused].Blur()
			m.addFocused = (m.addFocused + 1) % len(m.addInputs)
			m.addInputs[m.addFocused].Focus()
			return m, textinput.Blink

		case "shift+tab", "up":
			m.addInputs[m.addFocused].Blur()
			m.addFocused--
			if m.addFocused < 0 {
				m.addFocused = len(m.addInputs) - 1
			}
			m.addInputs[m.addFocused].Focus()
			return m, textinput.Blink

		case "enter":
			// Validate required fields
			name := m.addInputs[0].Value()
			username := m.addInputs[1].Value()
			password := m.addInputs[2].Value()

			if name == "" {
				m.err = fmt.Errorf("name is required")
				return m, nil
			}
			if password == "" {
				m.err = fmt.Errorf("password is required")
				return m, nil
			}

			// Create new entry and save to vault
			entry := store.Entry{
				Name:     name,
				Username: username,
				Password: password,
				URL:      m.addInputs[3].Value(),
				Notes:    m.addInputs[4].Value(),
			}

			syncResult, err := m.Vault.Add(entry)
			if err != nil {
				if err == store.ErrDuplicateEntry {
					m.err = fmt.Errorf("an entry with this name already exists")
				} else {
					m.err = fmt.Errorf("failed to save: %v", err)
				}
				return m, nil
			}

			// Refresh passwords from vault
			_ = m.refreshPasswords()

			m.err = nil
			m.view = ViewList
			return m, m.setToast(formatSyncToast("Saved", name, syncResult))
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	m.addInputs[m.addFocused], cmd = m.addInputs[m.addFocused].Update(msg)
	return m, cmd
}

func (m Model) viewAdd() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render("✚ Add New Password")

	b.WriteString(header)
	b.WriteString("\n\n")

	// Input fields
	labels := []string{"Name *", "Username", "Password *", "URL", "Notes"}
	for i, input := range m.addInputs {
		style := blurredStyle
		if i == m.addFocused {
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
