package main

import (
	"fmt"
	"os"

	"lockin/internal/store"
)

// Entry represents a credential to import
type Entry struct {
	Name     string
	Username string
	Password string
	URL      string // optional
	Notes    string // optional
}

func main() {
	masterPassword := ""

	entries := []Entry{
		// Example:
		// {Name: "github", Username: "user", Password: "pass123"},
	}

	if len(entries) == 0 {
		fmt.Println("No entries to import. Add entries to the 'entries' slice above.")
		os.Exit(0)
	}

	if masterPassword == "" {
		fmt.Println("Error: Please set your master password in the script.")
		os.Exit(1)
	}

	// Open vault
	vault, err := store.NewFileVault()
	if err != nil {
		fmt.Printf("Error opening vault: %v\n", err)
		os.Exit(1)
	}
	defer vault.Close()

	// Unlock vault with master password
	if err := vault.Unlock(masterPassword); err != nil {
		fmt.Printf("Error unlocking vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Importing %d entries...\n\n", len(entries))

	successCount := 0
	failCount := 0

	for _, e := range entries {
		storeEntry := store.Entry{
			Name:     e.Name,
			Username: e.Username,
			Password: e.Password,
			URL:      e.URL,
			Notes:    e.Notes,
		}

		if _, err := vault.Add(storeEntry); err != nil {
			fmt.Printf("✗ Failed to add '%s': %v\n", e.Name, err)
			failCount++
		} else {
			fmt.Printf("✓ Added '%s'\n", e.Name)
			successCount++
		}
	}

	fmt.Printf("\nImport complete: %d succeeded, %d failed\n", successCount, failCount)

	// Lock vault when done
	vault.Lock()
}
