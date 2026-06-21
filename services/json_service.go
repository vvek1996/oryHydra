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

