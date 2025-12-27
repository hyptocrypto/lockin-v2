package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			password := m.masterInput.Value()

			// Unlock the vault with the master password
			if err := m.Vault.Unlock(password); err != nil {
				m.err = err
				return m, nil
			}

			// Load passwords from the vault
			if err := m.refreshPasswords(); err != nil {
				// If decryption fails, the password was wrong
				m.err = fmt.Errorf("incorrect master password or corrupted data")
				m.Vault.Lock()
				return m, nil
			}

			m.err = nil
			m.view = ViewList
			m.masterInput.Reset()
			return m, nil

		case "esc":
			m.masterInput.Reset()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.masterInput, cmd = m.masterInput.Update(msg)
	return m, cmd
}

func (m Model) viewLogin() string {
	var b strings.Builder

	// Logo
	b.WriteString(logoStyle.Render(logo))
	b.WriteString("\n")

	// Title
	if m.isNewUser {
		b.WriteString(titleStyle.Render("Create Your Vault"))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Choose a strong master password to secure your passwords."))
	} else {
		b.WriteString(titleStyle.Render("Unlock Your Vault"))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("Enter your master password to access your passwords."))
	}
	b.WriteString("\n\n")

	// Password input
	b.WriteString(labelStyle.Render("Master Password"))
	b.WriteString("\n")
	b.WriteString(m.masterInput.View())
	b.WriteString("\n")

	// Error message
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("✗ %s", m.err.Error())))
	}

	// Help text
	if m.isNewUser {
		b.WriteString(helpStyle.Render("Press Enter to create vault • Ctrl+C to quit"))
	} else {
		b.WriteString(helpStyle.Render("Press Enter to unlock • Ctrl+C to quit"))
	}

	// Center the content
	content := boxStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
