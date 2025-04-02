package crypto

import (
	"testing"
)

func TestComputeHMAC256(t *testing.T) {
	tests := []struct {
		name       string
		toSign     []byte
		secretKey  string
		wantLength int
	}{
		{
			name:       "Basic HMAC test",
			toSign:     []byte("test data"),
			secretKey:  "secret key",
			wantLength: 64, // SHA-256 produces 32 bytes, which is 64 hex characters
		},
		{
			name:       "Empty data",
			toSign:     []byte(""),
			secretKey:  "secret key",
			wantLength: 64,
		},
		{
			name:       "Empty key",
			toSign:     []byte("test data"),
			secretKey:  "",
			wantLength: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeHMAC256(tt.toSign, tt.secretKey)
			if len(got) != tt.wantLength {
				t.Errorf("ComputeHMAC256() length = %v, want %v", len(got), tt.wantLength)
			}
		})
	}
}

func TestVerifyHMAC(t *testing.T) {
	tests := []struct {
		name                       string
		secretKey                  string
		toSign                     []byte
		providedSign               string
		compareOnlyFirstCharacters int
		want                       bool
	}{
		{
			name:                       "Valid signature",
			secretKey:                  "secret key",
			toSign:                     []byte("test data"),
			providedSign:               ComputeHMAC256([]byte("test data"), "secret key"),
			compareOnlyFirstCharacters: 0,
			want:                       true,
		},
		{
			name:                       "Invalid signature",
			secretKey:                  "secret key",
			toSign:                     []byte("test data"),
			providedSign:               "invalid signature",
			compareOnlyFirstCharacters: 0,
			want:                       false,
		},
		{
			name:                       "Compare first 8 characters - valid",
			secretKey:                  "secret key",
			toSign:                     []byte("test data"),
			providedSign:               ComputeHMAC256([]byte("test data"), "secret key"),
			compareOnlyFirstCharacters: 8,
			want:                       true,
		},
		{
			name:                       "Compare first 8 characters - invalid",
			secretKey:                  "secret key",
			toSign:                     []byte("test data"),
			providedSign:               "invalid" + ComputeHMAC256([]byte("test data"), "secret key")[8:],
			compareOnlyFirstCharacters: 8,
			want:                       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyHMAC(tt.secretKey, tt.toSign, tt.providedSign, tt.compareOnlyFirstCharacters); got != tt.want {
				t.Errorf("VerifyHMAC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "Long password",
			password: "this is a very long password with special characters !@#$%^&*()_+",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("HashPassword() returned empty string for valid password")
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password123"
	hash, _ := HashPassword(password)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "Valid password and hash",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "Invalid password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "Invalid hash",
			password: password,
			hash:     "invalidhash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPasswordHash(tt.password, tt.hash); got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha256Hash(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want int // expected length in bytes
	}{
		{
			name: "Basic string",
			str:  "test string",
			want: 32, // SHA-256 produces 32 bytes
		},
		{
			name: "Empty string",
			str:  "",
			want: 32,
		},
		{
			name: "Long string",
			str:  "this is a very long string that should still produce a 32-byte hash",
			want: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sha256Hash(tt.str)
			if len(got) != tt.want {
				t.Errorf("Sha256Hash() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestEncryptStringAndDecrypt(t *testing.T) {
	tests := []struct {
		name       string
		str        string
		passphrase string
		wantErr    bool
	}{
		{
			name:       "Basic encryption/decryption",
			str:        "test string",
			passphrase: "password123",
			wantErr:    false,
		},
		{
			name:       "Empty string",
			str:        "",
			passphrase: "password123",
			wantErr:    false,
		},
		{
			name:       "Long string",
			str:        "this is a very long string that should be encrypted and decrypted correctly",
			passphrase: "password123",
			wantErr:    false,
		},
		{
			name:       "Special characters",
			str:        "!@#$%^&*()_+{}|:\"<>?",
			passphrase: "password123",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test EncryptString
			encrypted, err := EncryptString(tt.str, tt.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && encrypted == "" {
				t.Error("EncryptString() returned empty string for valid input")
			}

			// Test DecryptFromHexString
			decrypted, err := DecryptFromHexString(encrypted, tt.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFromHexString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && decrypted != tt.str {
				t.Errorf("DecryptFromHexString() = %v, want %v", decrypted, tt.str)
			}
		})
	}
}

func TestDecryptFromHexString_Errors(t *testing.T) {
	tests := []struct {
		name       string
		str        string
		passphrase string
		wantErr    bool
	}{
		{
			name:       "Empty string",
			str:        "",
			passphrase: "password123",
			wantErr:    true,
		},
		{
			name:       "Invalid hex string",
			str:        "not a hex string",
			passphrase: "password123",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptFromHexString(tt.str, tt.passphrase)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFromHexString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
