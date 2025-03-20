package main

import (
	"encoding/base64"
	"fmt"

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
}
