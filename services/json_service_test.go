package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestFile(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_file.json")

	initialData := JSONData{
		Users: []User{
			{Email: "user1@example.com", Password: "pwd"},
			{Email: "user2@example.com", Password: "pwd"},
		},
		Groups: []string{"group1", "group2"},
		UserGroup: []UserGroup{
			{User: "user1@example.com", Group: "group1"},
		},
	}

	data, err := json.MarshalIndent(initialData, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal initial data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	return filePath
}

func TestAddGroup(t *testing.T) {
	filePath := setupTestFile(t)
	service, err := NewJsonFileService(filePath)
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	// 1. Successful addition
	err = service.AddGroup("group3")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify group list
	found := false
	for _, g := range service.JSONData.Groups {
		if g == "group3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected group3 to be added to Groups list")
	}

	// Verify saved state in file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	var savedData JSONData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatalf("failed to unmarshal test file: %v", err)
	}

	foundInFile := false
	for _, g := range savedData.Groups {
		if g == "group3" {
			foundInFile = true
			break
		}
	}
	if !foundInFile {
		t.Error("expected group3 to be persisted in file")
	}

	// 2. Duplicate group validation
	err = service.AddGroup("group3")
	if err == nil {
		t.Error("expected error when adding a duplicate group, got nil")
	}

	// 3. Empty group name validation
	err = service.AddGroup("")
	if err == nil {
		t.Error("expected error when adding empty group name, got nil")
	}
}

func TestAddUserToGroup(t *testing.T) {
	filePath := setupTestFile(t)
	service, err := NewJsonFileService(filePath)
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	// 1. Successful addition
	err = service.AddUserToGroup("user2@example.com", "group1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify addition in memory
	found := false
	for _, ug := range service.JSONData.UserGroup {
		if ug.User == "user2@example.com" && ug.Group == "group1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected user2 to be added to group1")
	}

	// Verify saved state in file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	var savedData JSONData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatalf("failed to unmarshal test file: %v", err)
	}

	foundInFile := false
	for _, ug := range savedData.UserGroup {
		if ug.User == "user2@example.com" && ug.Group == "group1" {
			foundInFile = true
			break
		}
	}
	if !foundInFile {
		t.Error("expected user2-group1 mapping to be persisted in file")
	}

	// 2. Duplicate group validation (same user, same group)
	err = service.AddUserToGroup("user2@example.com", "group1")
	if err == nil {
		t.Error("expected error when adding duplicate user to same group, got nil")
	}

	// 3. Non-existent user validation
	err = service.AddUserToGroup("nonexistent@example.com", "group1")
	if err == nil {
		t.Error("expected error when adding non-existent user, got nil")
	}

	// 4. Non-existent group validation
	err = service.AddUserToGroup("user2@example.com", "nonexistent_group")
	if err == nil {
		t.Error("expected error when adding to non-existent group, got nil")
	}
}

func TestRemoveGroup(t *testing.T) {
	filePath := setupTestFile(t)
	service, err := NewJsonFileService(filePath)
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	// 1. Successful removal (and cascading association removal)
	err = service.RemoveGroup("group1")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify group is removed in memory
	groupFound := false
	for _, g := range service.JSONData.Groups {
		if g == "group1" {
			groupFound = true
			break
		}
	}
	if groupFound {
		t.Error("expected group1 to be removed from Groups list")
	}

	// Verify user-group mapping for group1 is removed
	assocFound := false
	for _, ug := range service.JSONData.UserGroup {
		if ug.Group == "group1" {
			assocFound = true
			break
		}
	}
	if assocFound {
		t.Error("expected all user-group mappings for group1 to be removed")
	}

	// Verify saved state in file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	var savedData JSONData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatalf("failed to unmarshal test file: %v", err)
	}

	for _, g := range savedData.Groups {
		if g == "group1" {
			t.Error("expected group1 deletion to be persisted in file")
		}
	}
	for _, ug := range savedData.UserGroup {
		if ug.Group == "group1" {
			t.Error("expected group1 relation deletions to be persisted in file")
		}
	}

	// 2. Non-existent group removal validation
	err = service.RemoveGroup("nonexistent_group")
	if err == nil {
		t.Error("expected error when removing non-existent group, got nil")
	}
}
