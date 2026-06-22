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

func TestChangePassword(t *testing.T) {
	filePath := setupTestFile(t)
	service, err := NewJsonFileService(filePath)
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	// 1. Successful change
	err = service.ChangePassword("user1@example.com", "SecurePass123!")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify change in memory
	found := false
	for _, u := range service.JSONData.Users {
		if u.Email == "user1@example.com" {
			if u.Password != "SecurePass123!" {
				t.Errorf("expected password to be 'SecurePass123!', got %s", u.Password)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("expected user1@example.com to exist in Users list")
	}

	// Verify change is saved in file
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}
	var savedData JSONData
	if err := json.Unmarshal(content, &savedData); err != nil {
		t.Fatalf("failed to unmarshal test file: %v", err)
	}

	foundInFile := false
	for _, u := range savedData.Users {
		if u.Email == "user1@example.com" {
			if u.Password != "SecurePass123!" {
				t.Errorf("expected persisted password to be 'SecurePass123!', got %s", u.Password)
			}
			foundInFile = true
			break
		}
	}
	if !foundInFile {
		t.Error("expected user1@example.com to be persisted in file")
	}

	// 2. Non-existent user validation (using valid password format)
	err = service.ChangePassword("nonexistent@example.com", "ValidPass123!")
	if err == nil || err.Error() != "user not found" {
		t.Errorf("expected 'user not found' error, got %v", err)
	}

	// 3. Invalid password format validation
	err = service.ChangePassword("user1@example.com", "weak")
	if err == nil {
		t.Error("expected validation error for weak password, got nil")
	}

	// 4. Validation of empty email or password
	err = service.ChangePassword("", "ValidPass123!")
	if err == nil {
		t.Error("expected error when email is empty, got nil")
	}
	err = service.ChangePassword("user1@example.com", "")
	if err == nil {
		t.Error("expected error when password is empty, got nil")
	}
}

func TestRenameGroup(t *testing.T) {
	filePath := setupTestFile(t)
	service, err := NewJsonFileService(filePath)
	if err != nil {
		t.Fatalf("failed to initialize service: %v", err)
	}

	// 1. Successful rename and cascading mapping update
	err = service.RenameGroup("group1", "group1_renamed")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify rename in memory
	oldGroupFound := false
	newGroupFound := false
	for _, g := range service.JSONData.Groups {
		if g == "group1" {
			oldGroupFound = true
		}
		if g == "group1_renamed" {
			newGroupFound = true
		}
	}
	if oldGroupFound {
		t.Error("expected old group name 'group1' to be removed from Groups list")
	}
	if !newGroupFound {
		t.Error("expected new group name 'group1_renamed' to be added to Groups list")
	}

	// Verify userGroup relations are updated in memory
	relationUpdated := false
	for _, ug := range service.JSONData.UserGroup {
		if ug.Group == "group1" {
			t.Error("expected old group name 'group1' to be updated in all UserGroup associations")
		}
		if ug.User == "user1@example.com" && ug.Group == "group1_renamed" {
			relationUpdated = true
		}
	}
	if !relationUpdated {
		t.Error("expected user1 to be associated with 'group1_renamed'")
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
			t.Error("expected group1 rename to be persisted (old name should not exist)")
		}
	}
	foundNewGroupInFile := false
	for _, g := range savedData.Groups {
		if g == "group1_renamed" {
			foundNewGroupInFile = true
			break
		}
	}
	if !foundNewGroupInFile {
		t.Error("expected group1_renamed to be persisted in file")
	}

	foundNewRelationInFile := false
	for _, ug := range savedData.UserGroup {
		if ug.Group == "group1" {
			t.Error("expected old mapping for group1 to be updated in file")
		}
		if ug.User == "user1@example.com" && ug.Group == "group1_renamed" {
			foundNewRelationInFile = true
		}
	}
	if !foundNewRelationInFile {
		t.Error("expected updated user1-group1_renamed mapping to be persisted in file")
	}

	// 2. Non-existent old group name validation
	err = service.RenameGroup("nonexistent", "group3")
	if err == nil {
		t.Error("expected error when renaming non-existent group, got nil")
	}

	// 3. New name already exists validation
	err = service.RenameGroup("group1_renamed", "group2")
	if err == nil {
		t.Error("expected error when renaming to an already existing group name, got nil")
	}

	// 4. Empty names validation
	err = service.RenameGroup("", "newname")
	if err == nil {
		t.Error("expected error when old name is empty, got nil")
	}
	err = service.RenameGroup("group1_renamed", "")
	if err == nil {
		t.Error("expected error when new name is empty, got nil")
	}
}


