package services

import (
	"encoding/json"
	"fmt"
	"os"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserGroup struct {
	User  string `json:"user"`
	Group string `json:"group"`
}

type JSONData struct {
	Users     []User      `json:"user"`
	Groups    []string    `json:"group"`
	UserGroup []UserGroup `json:"userGroup"`
}

type JsonFileService struct {
	Path     string
	JSONData *JSONData
}

func NewJsonFileService(path string) (*JsonFileService, error) {
	service := &JsonFileService{
		Path: path,
	}

	if err := service.Read(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *JsonFileService) Read() error {
	content, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}

	var data JSONData
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	s.JSONData = &data
	return nil
}

// AddUser validates if the user already exists, checks password requirements, and appends them to the data store.
func (s *JsonFileService) AddUser(email string, password string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	email = TrimSpace(email)
	if email == "" || password == "" {
		return fmt.Errorf("email and password cannot be empty")
	}

	// Validate password strength
	if err := ValidatePassword(password); err != nil {
		return err
	}

	// Validate if user already exists
	for _, u := range s.JSONData.Users {
		if u.Email == email {
			return fmt.Errorf("user with email %s already exists", email)
		}
	}

	// Append new user
	s.JSONData.Users = append(s.JSONData.Users, User{
		Email:    email,
		Password: password,
	})

	// Save changes back to the JSON file
	return s.Save()
}

// ChangePassword updates a user's password after verifying password strength.
func (s *JsonFileService) ChangePassword(email string, newPassword string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	email = TrimSpace(email)
	if email == "" || newPassword == "" {
		return fmt.Errorf("email and password cannot be empty")
	}

	// Validate password strength
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	for i, u := range s.JSONData.Users {
		if u.Email == email {
			s.JSONData.Users[i].Password = newPassword
			return s.Save()
		}
	}

	return fmt.Errorf("user not found")
}

func (s *JsonFileService) GetUserGroups() []UserGroup {
	if s.JSONData == nil {
		return nil
	}

	return s.JSONData.UserGroup
}

func (s *JsonFileService) PrintUserGroup() {
	if s.JSONData == nil {
		return
	}
	for _, userGroup := range s.JSONData.UserGroup {
		fmt.Printf("user: %s, group: %s\n", userGroup.User, userGroup.Group)
	}
}

// Save writes JSONData back to the file.
func (s *JsonFileService) Save() error {
	if s.JSONData == nil {
		return fmt.Errorf("no data to save")
	}
	content, err := json.MarshalIndent(s.JSONData, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, content, 0644)
}

// AddGroup adds a group to s.JSONData.Groups if it does not exist.
func (s *JsonFileService) AddGroup(groupName string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	groupName = TrimSpace(groupName)
	if groupName == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	for _, g := range s.JSONData.Groups {
		if g == groupName {
			return fmt.Errorf("group already exists")
		}
	}
	s.JSONData.Groups = append(s.JSONData.Groups, groupName)
	return s.Save()
}

// AddUserToGroup validates user, group, and duplicate group relationship, then adds user to group.
func (s *JsonFileService) AddUserToGroup(userEmail string, groupName string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	userEmail = TrimSpace(userEmail)
	groupName = TrimSpace(groupName)
	// Validate user existence
	userExists := false
	for _, u := range s.JSONData.Users {
		if u.Email == userEmail {
			userExists = true
			break
		}
	}
	if !userExists {
		return fmt.Errorf("user not found")
	}

	// Validate group existence
	groupExists := false
	for _, g := range s.JSONData.Groups {
		if g == groupName {
			groupExists = true
			break
		}
	}
	if !groupExists {
		return fmt.Errorf("group not found")
	}

	// Validate duplicate group in same users
	for _, ug := range s.JSONData.UserGroup {
		if ug.User == userEmail && ug.Group == groupName {
			return fmt.Errorf("user is already in group")
		}
	}

	s.JSONData.UserGroup = append(s.JSONData.UserGroup, UserGroup{
		User:  userEmail,
		Group: groupName,
	})
	return s.Save()
}

// RemoveGroup removes group and its associations.
func (s *JsonFileService) RemoveGroup(groupName string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	groupName = TrimSpace(groupName)
	groupIndex := -1
	for i, g := range s.JSONData.Groups {
		if g == groupName {
			groupIndex = i
			break
		}
	}
	if groupIndex == -1 {
		return fmt.Errorf("group not found")
	}

	// Remove group from Groups list
	s.JSONData.Groups = append(s.JSONData.Groups[:groupIndex], s.JSONData.Groups[groupIndex+1:]...)

	// Clean up userGroup entries referring to this group
	var updatedUserGroups []UserGroup
	for _, ug := range s.JSONData.UserGroup {
		if ug.Group != groupName {
			updatedUserGroups = append(updatedUserGroups, ug)
		}
	}
	s.JSONData.UserGroup = updatedUserGroups

	return s.Save()
}

// RenameGroup renames an existing group and updates all of its user-group mappings.
func (s *JsonFileService) RenameGroup(oldName string, newName string) error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized")
	}
	oldName = TrimSpace(oldName)
	newName = TrimSpace(newName)
	if oldName == "" || newName == "" {
		return fmt.Errorf("group names cannot be empty")
	}

	// Verify old group exists and new group does not exist
	oldGroupIndex := -1
	for i, g := range s.JSONData.Groups {
		if g == oldName {
			oldGroupIndex = i
		}
		if g == newName {
			return fmt.Errorf("group with name %s already exists", newName)
		}
	}

	if oldGroupIndex == -1 {
		return fmt.Errorf("group to rename not found")
	}

	// Rename group in groups list
	s.JSONData.Groups[oldGroupIndex] = newName

	// Update userGroup relations
	for i, ug := range s.JSONData.UserGroup {
		if ug.Group == oldName {
			s.JSONData.UserGroup[i].Group = newName
		}
	}

	return s.Save()
}

// Validate checks data consistency, looking for duplicates and broken relations.
func (s *JsonFileService) Validate() error {
	if s.JSONData == nil {
		return fmt.Errorf("no data initialized to validate")
	}

	// 1. Validate Users (Check for duplicate emails)
	userMap := make(map[string]bool)
	for _, u := range s.JSONData.Users {
		if u.Email == "" {
			return fmt.Errorf("validation failed: found a user with an empty email")
		}
		if userMap[u.Email] {
			return fmt.Errorf("validation failed: duplicate user email found: %s", u.Email)
		}
		userMap[u.Email] = true
	}

	// 2. Validate Groups (Check for duplicate group names)
	groupMap := make(map[string]bool)
	for _, g := range s.JSONData.Groups {
		if g == "" {
			return fmt.Errorf("validation failed: found an empty group name")
		}
		if groupMap[g] {
			return fmt.Errorf("validation failed: duplicate group name found: %s", g)
		}
		groupMap[g] = true
	}

	// 3. Validate User-Group Relationships
	ugMap := make(map[string]bool)
	for _, ug := range s.JSONData.UserGroup {
		// Ensure the user in the relationship actually exists
		if !userMap[ug.User] {
			return fmt.Errorf("validation failed: user-group relation references non-existent user: %s", ug.User)
		}
		// Ensure the group in the relationship actually exists
		if !groupMap[ug.Group] {
			return fmt.Errorf("validation failed: user-group relation references non-existent group: %s", ug.Group)
		}

		// Check for duplicate user-group mappings
		relationKey := fmt.Sprintf("%s|%s", ug.User, ug.Group)
		if ugMap[relationKey] {
			return fmt.Errorf("validation failed: duplicate user-group relationship found for user %s in group %s", ug.User, ug.Group)
		}
		ugMap[relationKey] = true
	}

	return nil
}