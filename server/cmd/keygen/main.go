package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"aidanwoods.dev/go-paseto"
)

func main() {
	// Generate a new key pair
	secretKey := paseto.NewV4AsymmetricSecretKey()
	publicKey := secretKey.Public()

	// Convert keys to base64 for storage
	privateKeyBase64 := base64.StdEncoding.EncodeToString(secretKey.ExportBytes())
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey.ExportBytes())

	// Print the keys
	fmt.Printf("Generated PASETO v4 key pair:\n\n")
	fmt.Printf("Private Key (keep this secret!):\n%s\n\n", privateKeyBase64)
	fmt.Printf("Public Key:\n%s\n\n", publicKeyBase64)

	// If .env file exists, offer to update it
	if _, err := os.Stat(".env"); err == nil {
		fmt.Print("Would you like to update the .env file with these keys? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if response == "y" || response == "Y" {
			content, err := os.ReadFile(".env")
			if err != nil {
				log.Fatalf("Failed to read .env file: %v", err)
			}

			// Update the keys in the .env file
			updated := updateEnvContent(string(content), privateKeyBase64, publicKeyBase64)

			err = os.WriteFile(".env", []byte(updated), 0644)
			if err != nil {
				log.Fatalf("Failed to write .env file: %v", err)
			}
			fmt.Println("Updated .env file with new keys")
		}
	} else {
		fmt.Println("Note: Copy these values to your .env file:")
		fmt.Printf("PASETO_PRIVATE_KEY=%s\n", privateKeyBase64)
		fmt.Printf("PASETO_PUBLIC_KEY=%s\n", publicKeyBase64)
	}
}

func updateEnvContent(content, privateKey, publicKey string) string {
	var result string
	var foundPrivate, foundPublic bool

	// Split content into lines
	lines := make([]string, 0)
	current := ""
	for _, char := range content {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	// Update or append keys
	for _, line := range lines {
		if len(line) == 0 {
			result += line + "\n"
			continue
		}

		if line[0] == '#' {
			result += line + "\n"
			continue
		}

		if len(line) >= 18 && line[:18] == "PASETO_PRIVATE_KEY" {
			result += fmt.Sprintf("PASETO_PRIVATE_KEY=%s\n", privateKey)
			foundPrivate = true
		} else if len(line) >= 17 && line[:17] == "PASETO_PUBLIC_KEY" {
			result += fmt.Sprintf("PASETO_PUBLIC_KEY=%s\n", publicKey)
			foundPublic = true
		} else {
			result += line + "\n"
		}
	}

	// Append keys if they weren't found
	if !foundPrivate {
		result += fmt.Sprintf("PASETO_PRIVATE_KEY=%s\n", privateKey)
	}
	if !foundPublic {
		result += fmt.Sprintf("PASETO_PUBLIC_KEY=%s\n", publicKey)
	}

	return result
}
