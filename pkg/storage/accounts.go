package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DIDAccount represents a DID with its balance information
type DIDAccount struct {
	DID        string    `json:"did"`
	Balance    float64   `json:"balance"`
	DIDType    int       `json:"did_type"`
	PledgedRBT float64   `json:"pledged_rbt"`
	LockedRBT  float64   `json:"locked_rbt"`
	PinnedRBT  float64   `json:"pinned_rbt"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AccountsFile represents the structure of the accounts file
type AccountsFile struct {
	Version      string       `json:"version"`
	RubixNodeURL string       `json:"rubix_node_url"`
	ExportedAt   time.Time    `json:"exported_at"`
	TotalDIDs    int          `json:"total_dids"`
	Accounts     []DIDAccount `json:"accounts"`
}

// SaveAccountsToFile saves DID accounts to a JSON file
func SaveAccountsToFile(filepath string, accounts []DIDAccount, rubixNodeURL string) error {
	accountsFile := AccountsFile{
		Version:      "1.0",
		RubixNodeURL: rubixNodeURL,
		ExportedAt:   time.Now(),
		TotalDIDs:    len(accounts),
		Accounts:     accounts,
	}

	data, err := json.MarshalIndent(accountsFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadAccountsFromFile loads DID accounts from a JSON file
func LoadAccountsFromFile(filepath string) (*AccountsFile, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var accountsFile AccountsFile
	err = json.Unmarshal(data, &accountsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &accountsFile, nil
}

// FindAccountByDID finds an account by DID in the accounts file
func (af *AccountsFile) FindAccountByDID(did string) *DIDAccount {
	for _, account := range af.Accounts {
		if account.DID == did {
			return &account
		}
	}
	return nil
}

// GetAccountByIndex gets an account by its index (0-based)
func (af *AccountsFile) GetAccountByIndex(index int) *DIDAccount {
	if index < 0 || index >= len(af.Accounts) {
		return nil
	}
	return &af.Accounts[index]
}

// FilterByMinBalance returns accounts with balance >= minBalance
func (af *AccountsFile) FilterByMinBalance(minBalance float64) []DIDAccount {
	filtered := make([]DIDAccount, 0)
	for _, account := range af.Accounts {
		if account.Balance >= minBalance {
			filtered = append(filtered, account)
		}
	}
	return filtered
}
