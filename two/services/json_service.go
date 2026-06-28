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
	User  string   `json:"user"`
	Group []string `json:"group"`
}

type JSONData struct {
	Users     []User      `json:"user"`
	Groups    []string    `json:"group"`
	UserGroup []UserGroup `json:"userGroup"`
}

type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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

// Validate checks data consistency, checking for empty emails/groups, duplicate entries,
// weak passwords, and nonexistent mapped groups. It generates deduplicated errors and warnings.
func (s *JsonFileService) Validate() ([]ValidationError, []ValidationWarning) {
	var errors []ValidationError
	var warnings []ValidationWarning

	if s.JSONData == nil {
		return []ValidationError{{Code: "NIL_DATA", Message: "no data initialized to validate"}}, warnings
	}

	seenErrors := make(map[string]bool)
	seenWarnings := make(map[string]bool)

	addError := func(code, msg string) {
		key := fmt.Sprintf("%s:%s", code, msg)
		if !seenErrors[key] {
			seenErrors[key] = true
			errors = append(errors, ValidationError{Code: code, Message: msg})
		}
	}

	addWarning := func(code, msg string) {
		key := fmt.Sprintf("%s:%s", code, msg)
		if !seenWarnings[key] {
			seenWarnings[key] = true
			warnings = append(warnings, ValidationWarning{Code: code, Message: msg})
		}
	}

	reservedGroups := map[string]bool{
		"admin":  true,
		"viewer": true,
	}

	userEmails := s.validateUsers(addError)
	groupNames := s.validateGroups(addError, addWarning, reservedGroups)
	s.validateUserGroups(addError, addWarning, userEmails, groupNames, reservedGroups)

	return errors, warnings
}

func (s *JsonFileService) validateUsers(addError func(code, msg string)) map[string]bool {
	userEmails := make(map[string]bool)
	for _, u := range s.JSONData.Users {
		emailTrimmed := TrimSpace(u.Email)

		// email should not be empty
		if emailTrimmed == "" {
			addError("EMPTY_EMAIL", "email should not be empty")
			continue
		}

		// Validate email format
		if !IsValidEmail(emailTrimmed) {
			addError("INVALID_EMAIL", fmt.Sprintf("invalid email format: %s", emailTrimmed))
		}

		// user should not be duplicated
		if userEmails[emailTrimmed] {
			addError("DUPLICATE_USER", fmt.Sprintf("user %s should not be duplicated", emailTrimmed))
		}
		userEmails[emailTrimmed] = true

		// password validation needed
		if u.Password == "" {
			addError("EMPTY_PASSWORD", fmt.Sprintf("user %s password is wrong: password cannot be empty", emailTrimmed))
		} else {
			if err := ValidatePassword(u.Password); err != nil {
				addError("INVALID_PASSWORD", fmt.Sprintf("user %s password is wrong: %v", emailTrimmed, err))
			}
		}
	}
	return userEmails
}

func (s *JsonFileService) validateGroups(addError func(code, msg string), addWarning func(code, msg string), reservedGroups map[string]bool) map[string]bool {
	groupNames := make(map[string]bool)
	for _, g := range s.JSONData.Groups {
		gTrimmed := TrimSpace(g)

		// group should not be empty
		if gTrimmed == "" {
			addError("EMPTY_GROUP", "group name should not be empty")
			continue
		}

		// group should not be duplicated
		if groupNames[gTrimmed] {
			addError("DUPLICATE_GROUP", fmt.Sprintf("group %s should not be duplicated", gTrimmed))
		}
		groupNames[gTrimmed] = true

		// reserved group: admin, viewer -> if found need to save in warning
		if reservedGroups[gTrimmed] {
			addWarning("RESERVED_GROUP", fmt.Sprintf("group %s is a reserved group", gTrimmed))
		}
	}
	return groupNames
}

func (s *JsonFileService) validateUserGroups(addError func(code, msg string), addWarning func(code, msg string), userEmails map[string]bool, groupNames map[string]bool, reservedGroups map[string]bool) {
	ugUsers := make(map[string]bool)
	for _, ug := range s.JSONData.UserGroup {
		userTrimmed := TrimSpace(ug.User)

		if userTrimmed == "" {
			addError("EMPTY_EMAIL", "user email should not be empty in userGroup")
			continue
		}

		// Check if user exists in the user list
		if userTrimmed != "" && !userEmails[userTrimmed] {
			addError("USER_NOT_FOUND", fmt.Sprintf("user %s in userGroup is not defined in the user list", userTrimmed))
		}

		// user should not be duplicated in userGroup list
		if ugUsers[userTrimmed] {
			addError("DUPLICATE_USER", fmt.Sprintf("user %s should not be duplicated", userTrimmed))
		}
		ugUsers[userTrimmed] = true

		// group should not be empty (the group list itself)
		if len(ug.Group) == 0 {
			addError("EMPTY_GROUP", fmt.Sprintf("group list should not be empty for user %s", userTrimmed))
		}

		seenGroupInUser := make(map[string]bool)
		for _, g := range ug.Group {
			gTrimmed := TrimSpace(g)

			// group name inside the list should not be empty
			if gTrimmed == "" {
				addError("EMPTY_GROUP", fmt.Sprintf("group name should not be empty for user %s", userTrimmed))
				continue
			}

			// groups inside userGroups should be there in "group"
			if !groupNames[gTrimmed] {
				addError("GROUP_NOT_FOUND", fmt.Sprintf("group %s for user %s should be there in group list", gTrimmed, userTrimmed))
			}

			// group should not be duplicated inside a userGroup entry
			if seenGroupInUser[gTrimmed] {
				addError("DUPLICATE_GROUP", fmt.Sprintf("group %s should not be duplicated for user %s", gTrimmed, userTrimmed))
			}
			seenGroupInUser[gTrimmed] = true

			// reserved group: admin, viewer -> if found need to save in warning
			if reservedGroups[gTrimmed] {
				addWarning("RESERVED_GROUP", fmt.Sprintf("group %s is a reserved group", gTrimmed))
			}
		}
	}
}
