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

func main() {
	// 1. Read input plain text file
	plainBytes, err := os.ReadFile("test/intent.json")
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// 2. Load creation rules for a path that matches the .sops.yaml regex (e.g. ending in .enc.json)
	conf, err := config.LoadCreationRuleForFile(".sops.yaml", "test/intent.enc.json", nil)
	if err != nil {
		fmt.Printf("Error loading creation rules: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully loaded creation rules. KeyGroups count: %d\n", len(conf.KeyGroups))

	// 3. Load plain bytes using the appropriate store
	format := formats.FormatForPath("test/intent.json")
	storeConfig := config.NewStoresConfig()
	store := common.StoreForFormat(format, storeConfig)

	branches, err := store.LoadPlainFile(plainBytes)
	if err != nil {
		fmt.Printf("Error loading plain file: %v\n", err)
		os.Exit(1)
	}

	// 4. Create the Tree
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

	// 5. Generate DEK
	keyServices := []keyservice.KeyServiceClient{
		keyservice.NewLocalClient(),
	}
	dek, errs := tree.GenerateDataKeyWithKeyServices(keyServices)
	if len(errs) > 0 {
		fmt.Printf("Error generating DEK: %v\n", errs)
		os.Exit(1)
	}

	// 6. Encrypt the tree
	cipher := aes.NewCipher()
	mac, err := tree.Encrypt(dek, cipher)
	if err != nil {
		fmt.Printf("Error encrypting tree: %v\n", err)
		os.Exit(1)
	}

	// Encrypt the MAC itself
	encryptedMac, err := cipher.Encrypt(mac, dek, tree.Metadata.LastModified.Format(time.RFC3339))
	if err != nil {
		fmt.Printf("Error encrypting MAC: %v\n", err)
		os.Exit(1)
	}
	tree.Metadata.MessageAuthenticationCode = encryptedMac

	// 7. Emit the encrypted file
	encryptedBytes, err := store.EmitEncryptedFile(tree)
	if err != nil {
		fmt.Printf("Error emitting encrypted file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully encrypted JSON!")
	fmt.Println(string(encryptedBytes))

	// 8. Load age.key for decryption
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

	os.Setenv("SOPS_AGE_KEY", privateKey)

	// 9. Decrypt back
	decryptedBytes, err := decrypt.DataWithFormat(encryptedBytes, formats.Json)
	if err != nil {
		fmt.Printf("Error decrypting: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully decrypted JSON!")
	fmt.Println(string(decryptedBytes))
}
