package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/getsops/sops/v3"
	"github.com/getsops/sops/v3/aes"
	"github.com/getsops/sops/v3/cmd/sops/common"
	"github.com/getsops/sops/v3/cmd/sops/formats"
	"github.com/getsops/sops/v3/config"
	"github.com/getsops/sops/v3/decrypt"
	"github.com/getsops/sops/v3/keyservice"
)

// Encrypt encrypts the raw data using creation rules from configPath (e.g., ".sops.yaml")
// matched against matchingPath (e.g., "test/intent.enc.json").
func Encrypt(plainBytes []byte, configPath string, matchingPath string) ([]byte, error) {
	// Load creation rules
	conf, err := config.LoadCreationRuleForFile(configPath, matchingPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load creation rules: %w", err)
	}

	// Load plain bytes using the appropriate store
	format := formats.FormatForPath(matchingPath)
	storeConfig := config.NewStoresConfig()
	store := common.StoreForFormat(format, storeConfig)

	branches, err := store.LoadPlainFile(plainBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load plain file: %w", err)
	}

	// Create the Tree
	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups:               conf.KeyGroups,
			ShamirThreshold:         conf.ShamirThreshold,
			UnencryptedSuffix:       conf.UnencryptedSuffix,
			EncryptedSuffix:         conf.EncryptedSuffix,
			UnencryptedRegex:        conf.UnencryptedRegex,
			EncryptedRegex:          conf.EncryptedRegex,
			UnencryptedCommentRegex: conf.UnencryptedCommentRegex,
			EncryptedCommentRegex:   conf.EncryptedCommentRegex,
			MACOnlyEncrypted:        conf.MACOnlyEncrypted,
			LastModified:            time.Now().UTC(),
		},
	}

	// Generate DEK
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}
	dek, errs := tree.GenerateDataKeyWithKeyServices(keyServices)
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to generate DEK: %v", errs)
	}

	// Encrypt the tree
	cipher := aes.NewCipher()
	mac, err := tree.Encrypt(dek, cipher)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt tree: %w", err)
	}

	// Encrypt the MAC itself
	encryptedMac, err := cipher.Encrypt(mac, dek, tree.Metadata.LastModified.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt MAC: %w", err)
	}
	tree.Metadata.MessageAuthenticationCode = encryptedMac

	// Emit the encrypted file
	return store.EmitEncryptedFile(tree)
}

// Decrypt decrypts the encrypted payload using the provided age private key.
func Decrypt(encryptedBytes []byte, agePrivateKey string) ([]byte, error) {
	// Set the age key in memory for SOPS to discover
	os.Setenv("SOPS_AGE_KEY", agePrivateKey)

	// Decrypt back to raw data
	return decrypt.DataWithFormat(encryptedBytes, formats.Json)
}

func main() {
	// 1. Read input plain text file
	plainBytes, err := os.ReadFile("test/intent.json")
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// 2. Encrypt
	encryptedBytes, err := Encrypt(plainBytes, ".sops.yaml", "test/intent.enc.json")
	if err != nil {
		fmt.Printf("Error encrypting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully encrypted JSON!")
	fmt.Println(string(encryptedBytes))

	// 3. Load age.key for decryption
	ageKeyBytes, err := os.ReadFile("age.key")
	if err != nil {
		fmt.Printf("Error reading age.key: %v\n", err)
		os.Exit(1)
	}

	// Extract the secret key line
	var privateKey string
	for _, line := range strings.Split(string(ageKeyBytes), "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && trimmed[0] != '#' {
			privateKey = trimmed
			break
		}
	}
	if privateKey == "" {
		fmt.Println("No private key found in age.key")
		os.Exit(1)
	}

	// 4. Decrypt
	decryptedBytes, err := Decrypt(encryptedBytes, privateKey)
	if err != nil {
		fmt.Printf("Error decrypting: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully decrypted JSON!")
	fmt.Println(string(decryptedBytes))
}
