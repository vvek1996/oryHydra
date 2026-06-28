package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CLIResponse struct {
	Errors   []ValidationError   `json:"errors"`
	Warnings []ValidationWarning `json:"warnings"`
}

func getRepoPath() string {
	if envPath := os.Getenv("VALIDATION_REPO_PATH"); envPath != "" {
		return envPath
	}
	return "./repo"
}

func getValidationFilePath() string {
	// 1. Check environment variable
	if envPath := os.Getenv("VALIDATION_FILE_PATH"); envPath != "" {
		return envPath
	}
	// 2. Check local path relative to project 'one' directory (../two/file.json)
	if _, err := os.Stat("../two/file.json"); err == nil {
		return "../two/file.json"
	}
	// 3. Check current directory (file.json)
	if _, err := os.Stat("file.json"); err == nil {
		return "file.json"
	}
	// 4. Default path inside the cloned repository
	return filepath.Join(getRepoPath(), "file.json")
}

func ensureRepoCloned() error {
	repoPath := getRepoPath()
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		// Already cloned
		return nil
	}

	fmt.Println("Repository not found. Cloning...")
	_ = os.RemoveAll(repoPath)

	cmd := exec.Command("git", "clone", "--depth", "1", "-b", "file", "https://github.com/vvek1996/oryHydra.git", repoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %v, output: %s", err, string(output))
	}
	fmt.Println("Repository cloned successfully.")
	return nil
}

func pullRepo() (string, error) {
	repoPath := getRepoPath()
	cmd := exec.Command("git", "-C", repoPath, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git pull failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

func getGitCommit() string {
	repoPath := getRepoPath()
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
		return "N/A"
	}
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func main() {
	// Ensure the repository is cloned at startup
	if err := ensureRepoCloned(); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to initialize repository clone: %v\n", err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	r.GET("/validate", func(c *gin.Context) {
		filePath := getValidationFilePath()

		// Verify if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Validation file not found at: %s", filePath)})
			return
		}

		// Get validator path from env, default to local binary path
		validatorPath := os.Getenv("VALIDATOR_PATH")
		if validatorPath == "" {
			validatorPath = "./test1"
		}

		// Execute validator on the hardcoded file path
		cmd := exec.Command(validatorPath, "validate", filePath)
		output, err := cmd.Output()

		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to execute validator CLI: %v", err)})
				return
			}
		}

		// Unmarshal the validation output
		var cliResp CLIResponse
		if err := json.Unmarshal(output, &cliResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("Failed to parse validator output: %v", err),
				"output": string(output),
			})
			return
		}

		// Print result to console as requested
		if len(cliResp.Errors) > 0 {
			fmt.Printf("[VALIDATION ERROR] %d error(s) found in %s:\n", len(cliResp.Errors), filePath)
			for _, e := range cliResp.Errors {
				fmt.Printf("  - Code: %s, Message: %s\n", e.Code, e.Message)
			}
			c.JSON(http.StatusOK, gin.H{
				"status":    "error",
				"message":   "validation failed: errors are present",
				"commit_id": getGitCommit(),
				"details":   cliResp,
			})
		} else {
			fmt.Printf("[VALIDATION SUCCESS] No errors found in %s.\n", filePath)
			c.JSON(http.StatusOK, gin.H{
				"status":    "success",
				"message":   "validation succeeded: no errors found",
				"commit_id": getGitCommit(),
				"details":   cliResp,
			})
		}
	})

	// POST /pull - pulls the latest changes from the Git repository
	r.POST("/pull", func(c *gin.Context) {
		output, err := pullRepo()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to pull git repository updates",
				"error":   err.Error(),
			})
			return
		}

		// If broadcast query param is not set to "false", trigger peer-to-peer sync
		if c.Query("broadcast") != "false" {
			// DNS lookup for headless service to find all pod IPs
			ips, err := net.LookupHost("gin-server-headless")
			if err == nil {
				myIP := os.Getenv("POD_IP")
				for _, ip := range ips {
					if ip == myIP {
						continue // skip self
					}
					// Asynchronously call the other replicas to pull
					go func(targetIP string) {
						url := fmt.Sprintf("http://%s:4000/pull?broadcast=false", targetIP)
						resp, err := http.Post(url, "application/json", nil)
						if err == nil {
							resp.Body.Close()
						}
					}(ip)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Git repository updated successfully",
			"output":  output,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	fmt.Println("server running on port", port)
	r.Run(":" + port)
}
