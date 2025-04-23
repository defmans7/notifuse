package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAmazonSESValidation(t *testing.T) {
	passphrase := "test-passphrase"

	tests := []struct {
		name    string
		ses     AmazonSES
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty SES config",
			ses:     AmazonSES{},
			wantErr: false,
		},
		{
			name: "Valid SES config",
			ses: AmazonSES{
				Region:    "us-east-1",
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
				SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			wantErr: false,
		},
		{
			name: "Missing region",
			ses: AmazonSES{
				AccessKey: "AKIAIOSFODNN7EXAMPLE",
				SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			wantErr: true,
			errMsg:  "region is required",
		},
		{
			name: "Missing access key",
			ses: AmazonSES{
				Region:    "us-east-1",
				SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			},
			wantErr: true,
			errMsg:  "access key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ses.Validate(passphrase)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSMTPSettingsValidation(t *testing.T) {
	passphrase := "test-passphrase"

	tests := []struct {
		name    string
		smtp    SMTPSettings
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid SMTP settings",
			smtp: SMTPSettings{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user@example.com",
				Password: "password",
				UseTLS:   true,
			},
			wantErr: false,
		},
		{
			name: "Missing host",
			smtp: SMTPSettings{
				Port:     587,
				Username: "user@example.com",
				Password: "password",
			},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name: "Invalid port - zero",
			smtp: SMTPSettings{
				Host:     "smtp.example.com",
				Port:     0,
				Username: "user@example.com",
				Password: "password",
			},
			wantErr: true,
			errMsg:  "invalid port number",
		},
		{
			name: "Invalid port - too large",
			smtp: SMTPSettings{
				Host:     "smtp.example.com",
				Port:     70000,
				Username: "user@example.com",
				Password: "password",
			},
			wantErr: true,
			errMsg:  "invalid port number",
		},
		{
			name: "Missing username",
			smtp: SMTPSettings{
				Host:     "smtp.example.com",
				Port:     587,
				Password: "password",
			},
			wantErr: true,
			errMsg:  "username is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.smtp.Validate(passphrase)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSparkPostSettingsValidation(t *testing.T) {
	passphrase := "test-passphrase"

	tests := []struct {
		name      string
		sparkpost SparkPostSettings
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Valid SparkPost settings",
			sparkpost: SparkPostSettings{
				APIKey:      "test-api-key",
				Endpoint:    "https://api.sparkpost.com",
				SandboxMode: false,
			},
			wantErr: false,
		},
		{
			name: "Missing endpoint",
			sparkpost: SparkPostSettings{
				APIKey:      "test-api-key",
				SandboxMode: false,
			},
			wantErr: true,
			errMsg:  "endpoint is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sparkpost.Validate(passphrase)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEmailProviderValidation(t *testing.T) {
	passphrase := "test-passphrase"

	tests := []struct {
		name     string
		provider EmailProvider
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Empty provider",
			provider: EmailProvider{},
			wantErr:  false,
		},
		{
			name: "Valid SMTP provider",
			provider: EmailProvider{
				Kind:               EmailProviderKindSMTP,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
				SMTP: &SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user@example.com",
					Password: "password",
					UseTLS:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "Valid SES provider",
			provider: EmailProvider{
				Kind:               EmailProviderKindSES,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
				SES: &AmazonSES{
					Region:    "us-east-1",
					AccessKey: "AKIAIOSFODNN7EXAMPLE",
					SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid SparkPost provider",
			provider: EmailProvider{
				Kind:               EmailProviderKindSparkPost,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
				SparkPost: &SparkPostSettings{
					APIKey:   "test-api-key",
					Endpoint: "https://api.sparkpost.com",
				},
			},
			wantErr: false,
		},
		{
			name: "Missing default sender email",
			provider: EmailProvider{
				Kind:              EmailProviderKindSMTP,
				DefaultSenderName: "Default Sender",
				SMTP: &SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user@example.com",
					Password: "password",
				},
			},
			wantErr: true,
			errMsg:  "default sender email is required",
		},
		{
			name: "Invalid default sender email",
			provider: EmailProvider{
				Kind:               EmailProviderKindSMTP,
				DefaultSenderEmail: "invalid-email",
				DefaultSenderName:  "Default Sender",
				SMTP: &SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user@example.com",
					Password: "password",
				},
			},
			wantErr: true,
			errMsg:  "invalid default sender email",
		},
		{
			name: "Missing default sender name",
			provider: EmailProvider{
				Kind:               EmailProviderKindSMTP,
				DefaultSenderEmail: "default@example.com",
				SMTP: &SMTPSettings{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "user@example.com",
					Password: "password",
				},
			},
			wantErr: true,
			errMsg:  "default sender name is required",
		},
		{
			name: "Invalid kind",
			provider: EmailProvider{
				Kind:               "invalid",
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
			},
			wantErr: true,
			errMsg:  "invalid email provider kind",
		},
		{
			name: "SMTP provider with nil SMTP settings",
			provider: EmailProvider{
				Kind:               EmailProviderKindSMTP,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
			},
			wantErr: true,
			errMsg:  "SMTP settings required",
		},
		{
			name: "SES provider with nil SES settings",
			provider: EmailProvider{
				Kind:               EmailProviderKindSES,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
			},
			wantErr: true,
			errMsg:  "SES settings required",
		},
		{
			name: "SparkPost provider with nil SparkPost settings",
			provider: EmailProvider{
				Kind:               EmailProviderKindSparkPost,
				DefaultSenderEmail: "default@example.com",
				DefaultSenderName:  "Default Sender",
			},
			wantErr: true,
			errMsg:  "SparkPost settings required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.provider.Validate(passphrase)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEncryptionDecryption(t *testing.T) {
	passphrase := "test-passphrase"

	t.Run("SMTP password encryption/decryption", func(t *testing.T) {
		originalPassword := "test-password"
		smtp := SMTPSettings{
			Password: originalPassword,
		}

		// Encrypt
		err := smtp.EncryptPassword(passphrase)
		require.NoError(t, err)
		assert.NotEmpty(t, smtp.EncryptedPassword)
		assert.NotEqual(t, originalPassword, smtp.EncryptedPassword)

		// Clear password
		smtp.Password = ""

		// Decrypt
		err = smtp.DecryptPassword(passphrase)
		require.NoError(t, err)
		assert.Equal(t, originalPassword, smtp.Password)
	})

	t.Run("SES secret key encryption/decryption", func(t *testing.T) {
		originalSecretKey := "test-secret-key"
		ses := AmazonSES{
			SecretKey: originalSecretKey,
		}

		// Encrypt
		err := ses.EncryptSecretKey(passphrase)
		require.NoError(t, err)
		assert.NotEmpty(t, ses.EncryptedSecretKey)
		assert.NotEqual(t, originalSecretKey, ses.EncryptedSecretKey)

		// Clear secret key
		ses.SecretKey = ""

		// Decrypt
		err = ses.DecryptSecretKey(passphrase)
		require.NoError(t, err)
		assert.Equal(t, originalSecretKey, ses.SecretKey)
	})

	t.Run("SparkPost API key encryption/decryption", func(t *testing.T) {
		originalAPIKey := "test-api-key"
		sparkpost := SparkPostSettings{
			APIKey: originalAPIKey,
		}

		// Encrypt
		err := sparkpost.EncryptAPIKey(passphrase)
		require.NoError(t, err)
		assert.NotEmpty(t, sparkpost.EncryptedAPIKey)
		assert.NotEqual(t, originalAPIKey, sparkpost.EncryptedAPIKey)

		// Clear API key
		sparkpost.APIKey = ""

		// Decrypt
		err = sparkpost.DecryptAPIKey(passphrase)
		require.NoError(t, err)
		assert.Equal(t, originalAPIKey, sparkpost.APIKey)
	})
}

func TestEmailProviderEncryptDecryptSecretKeys(t *testing.T) {
	passphrase := "test-passphrase"

	t.Run("SMTP provider secret keys", func(t *testing.T) {
		provider := EmailProvider{
			Kind: EmailProviderKindSMTP,
			SMTP: &SMTPSettings{
				Password: "test-password",
			},
		}

		// Encrypt all secret keys
		err := provider.EncryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Empty(t, provider.SMTP.Password)
		assert.NotEmpty(t, provider.SMTP.EncryptedPassword)

		// Decrypt all secret keys
		err = provider.DecryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Equal(t, "test-password", provider.SMTP.Password)
	})

	t.Run("SES provider secret keys", func(t *testing.T) {
		provider := EmailProvider{
			Kind: EmailProviderKindSES,
			SES: &AmazonSES{
				SecretKey: "test-secret-key",
			},
		}

		// Encrypt all secret keys
		err := provider.EncryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Empty(t, provider.SES.SecretKey)
		assert.NotEmpty(t, provider.SES.EncryptedSecretKey)

		// Decrypt all secret keys
		err = provider.DecryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Equal(t, "test-secret-key", provider.SES.SecretKey)
	})

	t.Run("SparkPost provider secret keys", func(t *testing.T) {
		provider := EmailProvider{
			Kind: EmailProviderKindSparkPost,
			SparkPost: &SparkPostSettings{
				APIKey: "test-api-key",
			},
		}

		// Encrypt all secret keys
		err := provider.EncryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Empty(t, provider.SparkPost.APIKey)
		assert.NotEmpty(t, provider.SparkPost.EncryptedAPIKey)

		// Decrypt all secret keys
		err = provider.DecryptSecretKeys(passphrase)
		require.NoError(t, err)
		assert.Equal(t, "test-api-key", provider.SparkPost.APIKey)
	})
}
