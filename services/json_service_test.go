package services

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		data         JSONData
		wantErrors   []string // list of error codes expected
		wantWarnings []string // list of warning codes expected
	}{
		{
			name: "Valid Schema",
			data: JSONData{
				Users: []User{
					{Email: "test@example.com", Password: "SecurePass123!"},
				},
				Groups: []string{"group1"},
				UserGroup: []UserGroup{
					{User: "test@example.com", Group: []string{"group1"}},
				},
			},
			wantErrors:   nil,
			wantWarnings: nil,
		},
		{
			name: "Reserved group warnings",
			data: JSONData{
				Users: []User{
					{Email: "test@example.com", Password: "SecurePass123!"},
				},
				Groups: []string{"admin", "viewer"},
				UserGroup: []UserGroup{
					{User: "test@example.com", Group: []string{"admin"}},
				},
			},
			wantErrors:   nil,
			wantWarnings: []string{"RESERVED_GROUP", "RESERVED_GROUP"}, // admin and viewer top-level, and admin in mapping (deduplicated by code+message so 2 unique warnings)
		},
		{
			name: "Duplicate User and Duplicate Group in mappings",
			data: JSONData{
				Users: []User{
					{Email: "test@example.com", Password: "SecurePass123!"},
					{Email: "test@example.com", Password: "SecurePass123!"},
				},
				Groups: []string{"group1"},
				UserGroup: []UserGroup{
					{User: "test@example.com", Group: []string{"group1", "group1"}},
					{User: "test@example.com", Group: []string{"group1"}},
				},
			},
			wantErrors:   []string{"DUPLICATE_USER", "DUPLICATE_GROUP"},
			wantWarnings: nil,
		},
		{
			name: "Weak Password and Empty Group",
			data: JSONData{
				Users: []User{
					{Email: "test@example.com", Password: "weak"},
				},
				Groups: []string{"", "group1"},
				UserGroup: []UserGroup{
					{User: "test@example.com", Group: []string{"", "group2"}},
				},
			},
			wantErrors:   []string{"INVALID_PASSWORD", "EMPTY_GROUP", "EMPTY_GROUP", "GROUP_NOT_FOUND"},
			wantWarnings: nil,
		},
		{
			name: "Empty Email",
			data: JSONData{
				Users: []User{
					{Email: "", Password: "SecurePass123!"},
				},
			},
			wantErrors:   []string{"EMPTY_EMAIL"},
			wantWarnings: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &JsonFileService{
				JSONData: &tt.data,
			}
			errs, warns := service.Validate()

			if len(errs) != len(tt.wantErrors) {
				t.Errorf("got %d errors, want %d: %v", len(errs), len(tt.wantErrors), errs)
			} else {
				for i, errCode := range tt.wantErrors {
					if errs[i].Code != errCode {
						t.Errorf("error %d: got code %s, want %s (message: %s)", i, errs[i].Code, errCode, errs[i].Message)
					}
				}
			}

			if len(warns) != len(tt.wantWarnings) {
				t.Errorf("got %d warnings, want %d: %v", len(warns), len(tt.wantWarnings), warns)
			} else {
				for i, warnCode := range tt.wantWarnings {
					if warns[i].Code != warnCode {
						t.Errorf("warning %d: got code %s, want %s (message: %s)", i, warns[i].Code, warnCode, warns[i].Message)
					}
				}
			}
		})
	}
}
