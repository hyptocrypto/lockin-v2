package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If in search mode, handle search-specific input
	if m.searching {
		return m.updateSearch(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "shift+tab", "backtab":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "tab":
			if m.cursor < len(m.passwords)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.passwords) > 0 {
				m.selected = &m.passwords[m.cursor]
				m.view = ViewDetail
			}
		case "a":
			// Reset add form
			for i := range m.addInputs {
				m.addInputs[i].Reset()
			}
			m.addFocused = 0
			m.addInputs[0].Focus()
			m.view = ViewAdd
		case "/":
			// Enter search mode
			m.searching = true
			m.searchInput.Reset()
			m.searchInput.Focus()
			m.searchResults = m.passwords
			m.searchCursor = 0
			return m, textinput.Blink
		case "d":
			// Delete password with confirmation
			if len(m.passwords) > 0 && m.cursor < len(m.passwords) {
				entry := m.passwords[m.cursor]
				m.deleteTarget = &entry
				m.selected = &entry
				m.view = ViewConfirmDelete
			}
		case "r":
			// Refresh passwords from vault
			_ = m.refreshPasswords()
		case "q":
			// Lock vault and go back to login
			m.Vault.Lock()
			m.passwords = nil
			m.cursor = 0
			m.view = ViewLogin
		}
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Exit search mode
			m.searching = false
			m.searchInput.Reset()
			m.searchResults = nil
			return m, nil

		case "enter":
			// Select the current search result
			if len(m.searchResults) > 0 && m.searchCursor < len(m.searchResults) {
				m.selected = &m.searchResults[m.searchCursor]
				m.searching = false
				m.searchInput.Reset()
				m.searchResults = nil
				m.view = ViewDetail
			}
			return m, nil

		case "tab", "down", "ctrl+n":
			// Move to next result
			if m.searchCursor < len(m.searchResults)-1 {
				m.searchCursor++
			}
			return m, nil

		case "shift+tab", "backtab", "up", "ctrl+p":
			// Move to previous result
			if m.searchCursor > 0 {
				m.searchCursor--
			}
			return m, nil
		}
	}

	// Update search input
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)

	// Filter results based on current input
	m.searchResults = m.filterPasswords(m.searchInput.Value())
	if m.searchCursor >= len(m.searchResults) {
		m.searchCursor = max(0, len(m.searchResults)-1)
	}

	return m, cmd
}

// maxVisible is the maximum number of items to show at once
const maxVisible = 10

// getVisibleWindow calculates the start and end indices for a scrollable window
func getVisibleWindow(cursor, total, maxItems int) (start, end int) {
	if total <= maxItems {
		return 0, total
	}

	// Keep cursor centered when possible
	half := maxItems / 2
	start = cursor - half
	end = cursor + half + (maxItems % 2) // handle odd maxItems

	// Adjust if we're at the edges
	if start < 0 {
		start = 0
		end = maxItems
	}
	if end > total {
		end = total
		start = total - maxItems
	}

	return start, end
}

func (m Model) viewList() string {
	var b strings.Builder

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1).
		Render("🔐 Your Passwords")

	b.WriteString(header)
	b.WriteString("\n\n")

	// Search mode
	if m.searching {
		// Search input with autocomplete hint
		b.WriteString(focusedStyle.Render("Search: "))
		b.WriteString(m.searchInput.View())

		// Show autocomplete suggestion
		suggestion := m.getAutocompleteSuggestion()
		if suggestion != "" && suggestion != m.searchInput.Value() {
			remaining := suggestion[len(m.searchInput.Value()):]
			b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(remaining))
		}
		b.WriteString("\n\n")

		// Show filtered results with scrolling
		if len(m.searchResults) == 0 {
			emptyMsg := lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true).
				Render("No matches found")
			b.WriteString(emptyMsg)
		} else {
			start, end := getVisibleWindow(m.searchCursor, len(m.searchResults), maxVisible)

			// Show scroll indicator for items above
			if start > 0 {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("  ↑ %d more above\n", start)))
			}

			for i := start; i < end; i++ {
				entry := m.searchResults[i]
				cursor := "  "
				style := normalItemStyle
				if m.searchCursor == i {
					cursor = "▸ "
					style = selectedItemStyle
				}

				line := fmt.Sprintf("%s%s", cursor, entry.Name)
				if entry.Username != "" {
					line += fmt.Sprintf(" (%s)", entry.Username)
				}
				b.WriteString(style.Render(line))
				b.WriteString("\n")
			}

			// Show scroll indicator for items below
			if end < len(m.searchResults) {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("  ↓ %d more below", len(m.searchResults)-end)))
			}
		}

		// Help for search mode
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("Tab/↓ next • Shift+Tab/↑ prev • Enter select • Esc cancel"))
	} else {
		// Normal list view
		if len(m.passwords) == 0 {
			emptyMsg := lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true).
				Render("No passwords stored yet. Press 'a' to add one.")
			b.WriteString(emptyMsg)
		} else {
			start, end := getVisibleWindow(m.cursor, len(m.passwords), maxVisible)

			// Show scroll indicator for items above
			if start > 0 {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("  ↑ %d more above\n", start)))
			}

			for i := start; i < end; i++ {
				entry := m.passwords[i]
				cursor := "  "
				style := normalItemStyle
				if m.cursor == i {
					cursor = "▸ "
					style = selectedItemStyle
				}

				line := fmt.Sprintf("%s%s", cursor, entry.Name)
				if entry.Username != "" {
					line += fmt.Sprintf(" (%s)", entry.Username)
				}
				b.WriteString(style.Render(line))
				b.WriteString("\n")
			}

			// Show scroll indicator for items below
			if end < len(m.passwords) {
				b.WriteString(lipgloss.NewStyle().Foreground(mutedColor).Render(fmt.Sprintf("  ↓ %d more below", len(m.passwords)-end)))
			}
		}

		// Help
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("↑/↓ navigate • Enter select • / search • a add • d delete • q lock"))
	}

	// Center the content
	content := boxStyle.Width(60).Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
