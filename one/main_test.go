package main

import (
	"os"
	"testing"
)

func TestGetRepoPath_Default(t *testing.T) {
	os.Unsetenv("VALIDATION_REPO_PATH")
	expected := "./repo"
	actual := getRepoPath()
	if actual != expected {
		t.Errorf("Expected path %q, got %q", expected, actual)
	}
}

func TestGetRepoPath_Env(t *testing.T) {
	expected := "/tmp/custom_repo"
	os.Setenv("VALIDATION_REPO_PATH", expected)
	defer os.Unsetenv("VALIDATION_REPO_PATH")

	actual := getRepoPath()
	if actual != expected {
		t.Errorf("Expected path %q, got %q", expected, actual)
	}
}

func TestGetValidationFilePath_Env(t *testing.T) {
	expected := "/tmp/file.json"
	os.Setenv("VALIDATION_FILE_PATH", expected)
	defer os.Unsetenv("VALIDATION_FILE_PATH")

	actual := getValidationFilePath()
	if actual != expected {
		t.Errorf("Expected path %q, got %q", expected, actual)
	}
}

// func TestFailingExample_Invalid(t *testing.T) {
// 	// A deliberately failing test to demonstrate failure reporting in GitHub Actions
// 	t.Log("This test is expected to fail to show the failure formatting.")
// 	t.Fail()
// }

func TestSkippedExample(t *testing.T) {
	t.Skip("This test is skipped to demonstrate skip reporting.")
}
