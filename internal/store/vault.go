package store

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Errors
var (
	ErrVaultNotFound   = errors.New("vault not found")
	ErrInvalidPassword = errors.New("invalid master password")
	ErrVaultLocked     = errors.New("vault is locked")
	ErrEntryNotFound   = errors.New("entry not found")
	ErrDuplicateEntry  = errors.New("entry with this name already exists")
)

// Entry represents a password entry in the vault
type Entry struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	URL       string `json:"url,omitempty"`
	Notes     string `json:"notes,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// FileVault is a SQLite-based password vault with optional SMB sync
type FileVault struct {
	db            *sql.DB
	smb           *smbConnection
	obfuscatedKey []byte
	isUnlocked    bool
}

// NewFileVault creates and initializes a new vault
func NewFileVault() (*FileVault, error) {
	// Ensure config directory exists
	if err := os.MkdirAll(GetConfigDir(), 0700); err != nil {
		return nil, err
	}

	// Initialize logger
	if err := InitLogger(); err != nil {
		return nil, err
	}

	// Load configuration from YAML
	if err := LoadConfig(); err != nil {
		LogError("Failed to load config: %v", err)
		return nil, err
	}
	LogInfo("Config loaded, SMB enabled: %v", IsSMBEnabled())

	db, err := sql.Open("sqlite3", GetDBPath())
	if err != nil {
		LogError("Failed to open database: %v", err)
		return nil, err
	}
	LogInfo("Database opened: %s", GetDBPath())

	// Create table if needed
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS credentials (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			url TEXT,
			notes TEXT,
			created_at INTEGER,
			updated_at INTEGER
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	v := &FileVault{db: db}
	v.initSMB(!v.Exists())
	return v, nil
}

// Close closes the vault and all connections
func (v *FileVault) Close() error {
	LogInfo("Closing vault")
	v.closeSMB()
	CloseLogger()
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

// Unlock unlocks the vault with the master password
func (v *FileVault) Unlock(masterPassword string) error {
	key := deriveKey(masterPassword)
	v.setMasterKey(key)
	v.isUnlocked = true
	return nil
}

// Lock locks the vault and clears sensitive data
func (v *FileVault) Lock() {
	v.clearMasterKey()
	v.isUnlocked = false
	v.closeSMB()
}

// IsLocked returns whether the vault is locked
func (v *FileVault) IsLocked() bool {
	return !v.isUnlocked
}

// Exists checks if any credentials exist
func (v *FileVault) Exists() bool {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM credentials").Scan(&count)
	return err == nil && count > 0
}

// List returns all entries (decrypted)
func (v *FileVault) List() ([]Entry, error) {
	if v.IsLocked() {
		return nil, ErrVaultLocked
	}

	rows, err := v.db.Query(`
		SELECT id, name, username, password, url, notes, created_at, updated_at 
		FROM credentials ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return v.scanEntries(rows)
}

// Get retrieves an entry by ID
func (v *FileVault) Get(id int64) (*Entry, error) {
	if v.IsLocked() {
		return nil, ErrVaultLocked
	}

	row := v.db.QueryRow(`
		SELECT id, name, username, password, url, notes, created_at, updated_at 
		FROM credentials WHERE id = ?
	`, id)

	return v.scanEntry(row)
}

// GetByName retrieves an entry by name (case-insensitive)
func (v *FileVault) GetByName(name string) (*Entry, error) {
	if v.IsLocked() {
		return nil, ErrVaultLocked
	}

	row := v.db.QueryRow(`
		SELECT id, name, username, password, url, notes, created_at, updated_at 
		FROM credentials WHERE LOWER(name) = LOWER(?)
	`, name)

	return v.scanEntry(row)
}

// SyncResult contains the result of an operation with sync status
type SyncResult struct {
	SyncEnabled bool
	SyncError   error
}

// IsSyncEnabled returns true if SMB sync is enabled and connected
func (v *FileVault) IsSyncEnabled() bool {
	return IsSMBEnabled() && v.smb != nil
}

// Add adds a new entry
func (v *FileVault) Add(entry Entry) (SyncResult, error) {
	result := SyncResult{SyncEnabled: v.IsSyncEnabled()}

	if v.IsLocked() {
		return result, ErrVaultLocked
	}

	// Check for duplicate
	var count int
	if err := v.db.QueryRow("SELECT COUNT(*) FROM credentials WHERE LOWER(name) = LOWER(?)", entry.Name).Scan(&count); err != nil {
		return result, err
	}
	if count > 0 {
		return result, ErrDuplicateEntry
	}

	encUsername, err := v.encrypt(entry.Username)
	if err != nil {
		return result, err
	}
	encPassword, err := v.encrypt(entry.Password)
	if err != nil {
		return result, err
	}

	now := time.Now().Unix()
	_, err = v.db.Exec(`
		INSERT INTO credentials (name, username, password, url, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, entry.Name, encUsername, encPassword, entry.URL, entry.Notes, now, now)

	if err == nil {
		result.SyncError = v.Sync()
	}
	return result, err
}

// Update updates an existing entry
func (v *FileVault) Update(entry Entry) (SyncResult, error) {
	result := SyncResult{SyncEnabled: v.IsSyncEnabled()}

	if v.IsLocked() {
		return result, ErrVaultLocked
	}

	var count int
	if err := v.db.QueryRow("SELECT COUNT(*) FROM credentials WHERE id = ?", entry.ID).Scan(&count); err != nil {
		return result, err
	}
	if count == 0 {
		return result, ErrEntryNotFound
	}

	encUsername, err := v.encrypt(entry.Username)
	if err != nil {
		return result, err
	}
	encPassword, err := v.encrypt(entry.Password)
	if err != nil {
		return result, err
	}

	now := time.Now().Unix()
	_, err = v.db.Exec(`
		UPDATE credentials SET name=?, username=?, password=?, url=?, notes=?, updated_at=? WHERE id=?
	`, entry.Name, encUsername, encPassword, entry.URL, entry.Notes, now, entry.ID)

	if err == nil {
		result.SyncError = v.Sync()
	}
	return result, err
}

// Delete removes an entry by ID
func (v *FileVault) Delete(id int64) (SyncResult, error) {
	result := SyncResult{SyncEnabled: v.IsSyncEnabled()}

	if v.IsLocked() {
		return result, ErrVaultLocked
	}

	dbResult, err := v.db.Exec("DELETE FROM credentials WHERE id = ?", id)
	if err != nil {
		return result, err
	}

	rows, err := dbResult.RowsAffected()
	if err != nil {
		return result, err
	}
	if rows == 0 {
		return result, ErrEntryNotFound
	}

	result.SyncError = v.Sync()
	return result, nil
}

// Search searches entries by name
func (v *FileVault) Search(query string) ([]Entry, error) {
	if v.IsLocked() {
		return nil, ErrVaultLocked
	}

	rows, err := v.db.Query(`
		SELECT id, name, username, password, url, notes, created_at, updated_at 
		FROM credentials WHERE LOWER(name) LIKE LOWER(?) ORDER BY name ASC
	`, "%"+strings.TrimSpace(query)+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return v.scanEntries(rows)
}

// scanEntry scans a single row into an Entry (with decryption)
func (v *FileVault) scanEntry(row *sql.Row) (*Entry, error) {
	var e Entry
	var encUsername, encPassword string
	var url, notes sql.NullString
	var createdAt, updatedAt sql.NullInt64

	err := row.Scan(&e.ID, &e.Name, &encUsername, &encPassword, &url, &notes, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrEntryNotFound
	}
	if err != nil {
		return nil, err
	}

	e.Username, err = v.decrypt(encUsername)
	if err != nil {
		return nil, err
	}
	e.Password, err = v.decrypt(encPassword)
	if err != nil {
		return nil, err
	}

	if url.Valid {
		e.URL = url.String
	}
	if notes.Valid {
		e.Notes = notes.String
	}
	if createdAt.Valid {
		e.CreatedAt = createdAt.Int64
	}
	if updatedAt.Valid {
		e.UpdatedAt = updatedAt.Int64
	}

	return &e, nil
}

// scanEntries scans multiple rows into a slice of Entry
func (v *FileVault) scanEntries(rows *sql.Rows) ([]Entry, error) {
	var entries []Entry

	for rows.Next() {
		var e Entry
		var encUsername, encPassword string
		var url, notes sql.NullString
		var createdAt, updatedAt sql.NullInt64

		if err := rows.Scan(&e.ID, &e.Name, &encUsername, &encPassword, &url, &notes, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var err error
		e.Username, err = v.decrypt(encUsername)
		if err != nil {
			return nil, err
		}
		e.Password, err = v.decrypt(encPassword)
		if err != nil {
			return nil, err
		}

		if url.Valid {
			e.URL = url.String
		}
		if notes.Valid {
			e.Notes = notes.String
		}
		if createdAt.Valid {
			e.CreatedAt = createdAt.Int64
		}
		if updatedAt.Valid {
			e.UpdatedAt = updatedAt.Int64
		}

		entries = append(entries, e)
	}

	return entries, rows.Err()
}
