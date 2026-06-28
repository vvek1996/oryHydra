package main

import (
	"encoding/json"
	"fmt"
	"os"
	"test1/services"
)

type CLIResponse struct {
	Errors   []services.ValidationError   `json:"errors"`
	Warnings []services.ValidationWarning `json:"warnings"`
}

func main() {
	if len(os.Args) < 3 || os.Args[1] != "validate" {
		fmt.Fprintln(os.Stderr, "Usage: ./file-validate validate <filepath>")
		os.Exit(1)
	}

	filePath := os.Args[2]

	service, err := services.NewJsonFileService(filePath)
	if err != nil {
		resp := CLIResponse{
			Errors: []services.ValidationError{
				{
					Code:    "INVALID_JSON",
					Message: fmt.Sprintf("failed to parse or read JSON file: %v", err),
				},
			},
			Warnings: []services.ValidationWarning{},
		}
		printJSONAndExit(resp, 1)
	}

	errors, warnings := service.Validate()

	if errors == nil {
		errors = []services.ValidationError{}
	}
	if warnings == nil {
		warnings = []services.ValidationWarning{}
	}

	resp := CLIResponse{
		Errors:   errors,
		Warnings: warnings,
	}

	exitCode := 0
	if len(errors) > 0 {
		exitCode = 1
	}

	printJSONAndExit(resp, exitCode)
}

func printJSONAndExit(resp CLIResponse, exitCode int) {
	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal response to JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(bytes))
	os.Exit(exitCode)
}
