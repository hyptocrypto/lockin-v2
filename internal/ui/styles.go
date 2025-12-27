package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	accentColor    = lipgloss.Color("#F59E0B") // Amber
	errorColor     = lipgloss.Color("#EF4444") // Red
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	textColor      = lipgloss.Color("#F9FAFB") // Light text

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// Subtitle/description style
	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginBottom(2)

	// Input label style
	labelStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Bold(true)

	// Focused input style
	focusedStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	// Blurred/unfocused input style
	blurredStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Error message style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success message style
	successStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	// Help text style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(2)

	// Box/container style
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Logo/header style
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(2).Align(lipgloss.Center)

	// Toast notification style
	toastStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true).
			Padding(0, 1).
			MarginTop(1)
)

// Logo for the application
const logo = `
    __               __   ____    
   / /   ____  _____/ /__/  _/___ 
  / /   / __ \/ ___/ //_// // __ \
 / /___/ /_/ / /__/ ,< _/ // / / /
/_____/\____/\___/_/|_/___/_/ /_/ 
                                                                   
                                 `
