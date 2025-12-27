package ui

import (
	"lockin/internal/store"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

// ToastClearMsg is sent when the toast should be cleared
type ToastClearMsg struct{}

// toastDuration is how long the toast is displayed
const toastDuration = 3 * time.Second

// clearToastAfter returns a command that clears the toast after a delay
func clearToastAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return ToastClearMsg{}
	})
}

// setToast sets the toast message and returns a command to clear it
func (m *Model) setToast(text string) tea.Cmd {
	m.toastText = text
	return clearToastAfter(toastDuration)
}

// formatSyncToast formats a toast message with sync status
func formatSyncToast(action, name string, sync store.SyncResult) string {
	msg := "✓ " + action + " '" + name + "'"
	if sync.SyncEnabled {
		if sync.SyncError != nil {
			msg += " (sync failed)"
		} else {
			msg += " (synced)"
		}
	}
	return msg
}

// View represents the current screen/view in the application
type View int

const (
	ViewLogin View = iota
	ViewUnlock
	ViewList
	ViewAdd
	ViewDetail
	ViewEdit
	ViewConfirmDelete
)

// Model is the main application model
type Model struct {
	view   View
	width  int
	height int
	err    error

	// Login/Unlock state
	masterInput textinput.Model
	isNewUser   bool

	// Password list state
	passwords []PasswordEntry
	cursor    int

	// Search state
	searching     bool
	searchInput   textinput.Model
	searchResults []PasswordEntry
	searchCursor  int

	// Add password state
	addInputs  []textinput.Model
	addFocused int

	// Edit password state
	editInputs  []textinput.Model
	editFocused int
	editingID   int64

	// Selected password for detail view
	selected *PasswordEntry

	// Delete confirmation state
	confirmingDelete bool
	deleteTarget     *PasswordEntry

	// Storage
	Vault *store.FileVault

	// Notification
	toastText string
}

// PasswordEntry represents a stored password (UI representation)
type PasswordEntry struct {
	ID       int64
	Name     string
	Username string
	Password string
	URL      string
	Notes    string
}

// ToStoreEntry converts a PasswordEntry to a store.Entry
func (p PasswordEntry) ToStoreEntry() store.Entry {
	return store.Entry{
		ID:       p.ID,
		Name:     p.Name,
		Username: p.Username,
		Password: p.Password,
		URL:      p.URL,
		Notes:    p.Notes,
	}
}

// FromStoreEntry creates a PasswordEntry from a store.Entry
func FromStoreEntry(e store.Entry) PasswordEntry {
	return PasswordEntry{
		ID:       e.ID,
		Name:     e.Name,
		Username: e.Username,
		Password: e.Password,
		URL:      e.URL,
		Notes:    e.Notes,
	}
}

// New creates and returns a new Model with initial state
func New() Model {
	// Master password input
	masterInput := textinput.New()
	masterInput.Placeholder = "Enter master password..."
	masterInput.Focus()
	masterInput.EchoMode = textinput.EchoPassword
	masterInput.EchoCharacter = '•'
	masterInput.CharLimit = 128
	masterInput.Width = 40

	// Add password form inputs
	addInputs := make([]textinput.Model, 5)
	placeholders := []string{"Name", "Username", "Password", "URL (optional)", "Notes (optional)"}
	for i := range addInputs {
		addInputs[i] = textinput.New()
		addInputs[i].Placeholder = placeholders[i]
		addInputs[i].CharLimit = 256
		addInputs[i].Width = 40
		if i == 2 { // Password field
			addInputs[i].EchoMode = textinput.EchoPassword
			addInputs[i].EchoCharacter = '•'
		}
	}
	addInputs[0].Focus()

	// Search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search..."
	searchInput.CharLimit = 128
	searchInput.Width = 40

	// Edit password form inputs (same structure as add)
	editInputs := make([]textinput.Model, 5)
	for i := range editInputs {
		editInputs[i] = textinput.New()
		editInputs[i].Placeholder = placeholders[i]
		editInputs[i].CharLimit = 256
		editInputs[i].Width = 40
		if i == 2 { // Password field
			editInputs[i].EchoMode = textinput.EchoPassword
			editInputs[i].EchoCharacter = '•'
		}
	}

	vault, err := store.NewFileVault()
	if err != nil {
		panic("failed to open vault: " + err.Error())
	}

	// Check if this is a new user (no existing credentials)
	isNewUser := !vault.Exists()

	return Model{
		view:          ViewLogin,
		isNewUser:     isNewUser,
		masterInput:   masterInput,
		addInputs:     addInputs,
		editInputs:    editInputs,
		searchInput:   searchInput,
		passwords:     []PasswordEntry{},
		searchResults: []PasswordEntry{},
		Vault:         vault,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ToastClearMsg:
		m.toastText = ""
		return m, nil
	}

	// Delegate to view-specific update handlers
	switch m.view {
	case ViewLogin, ViewUnlock:
		return m.updateLogin(msg)
	case ViewList:
		return m.updateList(msg)
	case ViewAdd:
		return m.updateAdd(msg)
	case ViewDetail:
		return m.updateDetail(msg)
	case ViewEdit:
		return m.updateEdit(msg)
	case ViewConfirmDelete:
		return m.updateConfirmDelete(msg)
	}

	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	var content string
	switch m.view {
	case ViewLogin, ViewUnlock:
		content = m.viewLogin()
	case ViewList:
		content = m.viewList()
	case ViewAdd:
		content = m.viewAdd()
	case ViewDetail:
		content = m.viewDetail()
	case ViewEdit:
		content = m.viewEdit()
	case ViewConfirmDelete:
		content = m.viewConfirmDelete()
	default:
		content = "Unknown view"
	}
	return m.withToastOverlay(content)
}

// refreshPasswords loads passwords from the vault into the model
func (m *Model) refreshPasswords() error {
	entries, err := m.Vault.List()
	if err != nil {
		return err
	}

	m.passwords = make([]PasswordEntry, len(entries))
	for i, e := range entries {
		m.passwords[i] = FromStoreEntry(e)
	}
	return nil
}

// filterPasswords returns passwords that match the search query (case-insensitive)
func (m *Model) filterPasswords(query string) []PasswordEntry {
	if query == "" {
		return m.passwords
	}

	query = strings.ToLower(query)
	var results []PasswordEntry
	for _, p := range m.passwords {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Username), query) {
			results = append(results, p)
		}
	}
	return results
}

// getAutocompleteSuggestion returns the best autocomplete suggestion for the current query
func (m *Model) getAutocompleteSuggestion() string {
	query := m.searchInput.Value()
	if query == "" {
		return ""
	}

	queryLower := strings.ToLower(query)
	for _, p := range m.passwords {
		nameLower := strings.ToLower(p.Name)
		if strings.HasPrefix(nameLower, queryLower) {
			return p.Name
		}
	}
	return ""
}

func (m Model) withToastOverlay(bg string) string {
	if m.toastText != "" {
		toast := toastStyle.Render(m.toastText)
		return overlay.Composite(
			toast,
			bg,
			overlay.Right,
			overlay.Top,
			-2,
			-1,
		)
	}

	return bg
}
