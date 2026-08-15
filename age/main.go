package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"filippo.io/age"
)

func main() {
	// 1. Generate X25519 Key Pair
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		log.Fatalf("Failed to generate identity: %v", err)
	}

	publicKey := identity.Recipient().String()
	privateKey := identity.String()

	// 2. Write key.txt (Standard age-keygen format)
	keyFileContent := fmt.Sprintf(
		"# created: %s\n# public key: %s\n%s\n",
		time.Now().Format(time.RFC3339),
		publicKey,
		privateKey,
	)

	err = os.WriteFile("age.key", []byte(keyFileContent), 0600) // 0600 = readable only by user
	if err != nil {
		log.Fatalf("Failed to write age.key: %v", err)
	}
	fmt.Println("Successfully generated: age.key")

	// 3. Write .sops.yaml
	sopsYamlContent := fmt.Sprintf(`creation_rules:
  - path_regex: \.enc\.json$
    age: "%s"
`, publicKey)

	err = os.WriteFile(".sops.yaml", []byte(sopsYamlContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write .sops.yaml: %v", err)
	}
	fmt.Println("Successfully generated: .sops.yaml")
}
