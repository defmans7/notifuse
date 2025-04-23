package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/pkg/crypto"
	"github.com/asaskevich/govalidator"
)

//go:generate mockgen -destination mocks/mock_workspace_repository.go -package mocks github.com/Notifuse/notifuse/internal/domain WorkspaceRepository
//go:generate mockgen -destination mocks/mock_workspace_service.go -package mocks github.com/Notifuse/notifuse/internal/domain WorkspaceServiceInterface

// WorkspaceSettings contains configurable workspace settings
type WorkspaceSettings struct {
	WebsiteURL                 string              `json:"website_url,omitempty"`
	LogoURL                    string              `json:"logo_url,omitempty"`
	CoverURL                   string              `json:"cover_url,omitempty"`
	Timezone                   string              `json:"timezone"`
	FileManager                FileManagerSettings `json:"file_manager,omitempty"`
	EmailMarketingProvider     EmailProvider       `json:"email_marketing,omitempty"`
	EmailTransactionalProvider EmailProvider       `json:"email_transactional,omitempty"`
}

// Validate validates workspace settings
func (ws *WorkspaceSettings) Validate(passphrase string) error {
	if ws.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}

	if !IsValidTimezone(ws.Timezone) {
		return fmt.Errorf("invalid timezone: %s", ws.Timezone)
	}

	if ws.WebsiteURL != "" && !govalidator.IsURL(ws.WebsiteURL) {
		return fmt.Errorf("invalid website URL: %s", ws.WebsiteURL)
	}

	if ws.LogoURL != "" && !govalidator.IsURL(ws.LogoURL) {
		return fmt.Errorf("invalid logo URL: %s", ws.LogoURL)
	}

	if ws.CoverURL != "" && !govalidator.IsURL(ws.CoverURL) {
		return fmt.Errorf("invalid cover URL: %s", ws.CoverURL)
	}

	// FileManager is completely optional, but if any fields are set, validate them
	if err := ws.FileManager.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid file manager settings: %w", err)
	}

	// Email providers are completely optional, but if they are configured, validate them
	if err := ws.EmailMarketingProvider.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid email marketing provider settings: %w", err)
	}

	if err := ws.EmailTransactionalProvider.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid email transactional provider settings: %w", err)
	}

	return nil
}

type Workspace struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Settings  WorkspaceSettings `json:"settings"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Validate performs validation on the workspace fields
func (w *Workspace) Validate(passphrase string) error {
	// Validate ID
	if w.ID == "" {
		return fmt.Errorf("invalid workspace: id is required")
	}
	if !govalidator.IsAlphanumeric(w.ID) {
		return fmt.Errorf("invalid workspace: id must be alphanumeric")
	}
	if len(w.ID) > 32 {
		return fmt.Errorf("invalid workspace: id length must be between 1 and 32")
	}

	// Validate Name
	if w.Name == "" {
		return fmt.Errorf("invalid workspace: name is required")
	}
	if len(w.Name) > 255 {
		return fmt.Errorf("invalid workspace: name length must be between 1 and 255")
	}

	// Validate Settings
	if err := w.Settings.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid workspace settings: %w", err)
	}

	return nil
}

func (w *Workspace) BeforeSave(secretkey string) error {
	// Only process FileManager if there's a SecretKey to encrypt
	if w.Settings.FileManager.SecretKey != "" {
		if err := w.Settings.FileManager.EncryptSecretKey(secretkey); err != nil {
			return fmt.Errorf("failed to encrypt secret key: %w", err)
		}
		// clear the secret key from the workspace settings
		w.Settings.FileManager.SecretKey = ""
	}

	// Encrypt secrets for email providers
	if err := w.Settings.EmailMarketingProvider.EncryptSecretKeys(secretkey); err != nil {
		return fmt.Errorf("failed to encrypt email marketing provider secrets: %w", err)
	}

	if err := w.Settings.EmailTransactionalProvider.EncryptSecretKeys(secretkey); err != nil {
		return fmt.Errorf("failed to encrypt email transactional provider secrets: %w", err)
	}

	return nil
}

func (w *Workspace) AfterLoad(secretkey string) error {
	// Only decrypt if there's an EncryptedSecretKey present
	if w.Settings.FileManager.EncryptedSecretKey != "" {
		if err := w.Settings.FileManager.DecryptSecretKey(secretkey); err != nil {
			return fmt.Errorf("failed to decrypt secret key: %w", err)
		}
	}

	// Decrypt secrets for email providers
	if err := w.Settings.EmailMarketingProvider.DecryptSecretKeys(secretkey); err != nil {
		return fmt.Errorf("failed to decrypt email marketing provider secrets: %w", err)
	}

	if err := w.Settings.EmailTransactionalProvider.DecryptSecretKeys(secretkey); err != nil {
		return fmt.Errorf("failed to decrypt email transactional provider secrets: %w", err)
	}

	return nil
}

type FileManagerSettings struct {
	Endpoint           string  `json:"endpoint"`
	Bucket             string  `json:"bucket"`
	AccessKey          string  `json:"access_key"`
	EncryptedSecretKey string  `json:"encrypted_secret_key,omitempty"`
	Region             *string `json:"region,omitempty"`
	CDNEndpoint        *string `json:"cdn_endpoint,omitempty"`

	// decoded secret key, not stored in the database
	SecretKey string `json:"secret_key,omitempty"`
}

func (f *FileManagerSettings) DecryptSecretKey(passphrase string) error {
	secretKey, err := crypto.DecryptFromHexString(f.EncryptedSecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt secret key: %w", err)
	}
	f.SecretKey = secretKey
	return nil
}

func (f *FileManagerSettings) EncryptSecretKey(passphrase string) error {
	encryptedSecretKey, err := crypto.EncryptString(f.SecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret key: %w", err)
	}
	f.EncryptedSecretKey = encryptedSecretKey
	return nil
}

func (f *FileManagerSettings) Validate(passphrase string) error {
	// Check if any field is set to determine if we should validate
	isConfigured := f.Endpoint != "" || f.Bucket != "" || f.AccessKey != "" ||
		f.EncryptedSecretKey != "" || f.SecretKey != "" ||
		(f.Region != nil) || (f.CDNEndpoint != nil)

	// If no fields are set, consider it valid (optional config)
	if !isConfigured {
		return nil
	}

	// If any field is set, validate required fields are present
	if f.Endpoint == "" {
		return fmt.Errorf("endpoint is required when file manager is configured")
	}

	if !govalidator.IsURL(f.Endpoint) {
		return fmt.Errorf("invalid endpoint: %s", f.Endpoint)
	}

	if f.Bucket == "" {
		return fmt.Errorf("bucket is required when file manager is configured")
	}

	if f.AccessKey == "" {
		return fmt.Errorf("access key is required when file manager is configured")
	}

	// Region is optional, so we don't check if it's empty
	if f.CDNEndpoint != nil && !govalidator.IsURL(*f.CDNEndpoint) {
		return fmt.Errorf("invalid cdn endpoint: %s", *f.CDNEndpoint)
	}

	// only encrypt secret key if it's not empty
	if f.SecretKey != "" {
		if err := f.EncryptSecretKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt secret key: %w", err)
		}
	}

	return nil
}

// AmazonSES contains SES email provider settings
type AmazonSES struct {
	Region             string `json:"region"`
	AccessKey          string `json:"access_key"`
	EncryptedSecretKey string `json:"encrypted_secret_key,omitempty"`
	SenderEmail        string `json:"sender_email"`
	SandboxMode        bool   `json:"sandbox_mode"`

	// decoded secret key, not stored in the database
	SecretKey string `json:"secret_key,omitempty"`
}

func (a *AmazonSES) DecryptSecretKey(passphrase string) error {
	secretKey, err := crypto.DecryptFromHexString(a.EncryptedSecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SES secret key: %w", err)
	}
	a.SecretKey = secretKey
	return nil
}

func (a *AmazonSES) EncryptSecretKey(passphrase string) error {
	encryptedSecretKey, err := crypto.EncryptString(a.SecretKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SES secret key: %w", err)
	}
	a.EncryptedSecretKey = encryptedSecretKey
	return nil
}

func (a *AmazonSES) Validate(passphrase string) error {
	// Check if any field is set to determine if we should validate
	isConfigured := a.Region != "" || a.AccessKey != "" ||
		a.EncryptedSecretKey != "" || a.SecretKey != "" || a.SenderEmail != ""

	// If no fields are set, consider it valid (optional config)
	if !isConfigured {
		return nil
	}

	// If any field is set, validate required fields are present
	if a.Region == "" {
		return fmt.Errorf("region is required when Amazon SES is configured")
	}

	if a.AccessKey == "" {
		return fmt.Errorf("access key is required when Amazon SES is configured")
	}

	if a.SenderEmail == "" {
		return fmt.Errorf("sender email is required when Amazon SES is configured")
	}

	if !govalidator.IsEmail(a.SenderEmail) {
		return fmt.Errorf("invalid sender email: %s", a.SenderEmail)
	}

	// only encrypt secret key if it's not empty
	if a.SecretKey != "" {
		if err := a.EncryptSecretKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SES secret key: %w", err)
		}
	}

	return nil
}

// EmailProviderKind defines the type of email provider
type EmailProviderKind string

const (
	EmailProviderKindSMTP      EmailProviderKind = "smtp"
	EmailProviderKindSES       EmailProviderKind = "ses"
	EmailProviderKindSparkPost EmailProviderKind = "sparkpost"
)

// EmailProvider contains configuration for an email service provider
type EmailProvider struct {
	Kind      EmailProviderKind  `json:"kind"`
	SES       *AmazonSES         `json:"ses,omitempty"`
	SMTP      *SMTPSettings      `json:"smtp,omitempty"`
	SparkPost *SparkPostSettings `json:"sparkpost,omitempty"`
}

// Validate validates the email provider settings
func (e *EmailProvider) Validate(passphrase string) error {
	// If Kind is empty, consider it as not configured
	if e.Kind == "" {
		return nil
	}

	// Validate Kind value
	switch e.Kind {
	case EmailProviderKindSMTP:
		if e.SMTP == nil {
			return fmt.Errorf("SMTP settings required when email provider kind is smtp")
		}
		return e.SMTP.Validate(passphrase)
	case EmailProviderKindSES:
		if e.SES == nil {
			return fmt.Errorf("SES settings required when email provider kind is ses")
		}
		return e.SES.Validate(passphrase)
	case EmailProviderKindSparkPost:
		if e.SparkPost == nil {
			return fmt.Errorf("SparkPost settings required when email provider kind is sparkpost")
		}
		return e.SparkPost.Validate(passphrase)
	default:
		return fmt.Errorf("invalid email provider kind: %s", e.Kind)
	}
}

// EncryptSecretKeys encrypts all secret keys in the email provider
func (e *EmailProvider) EncryptSecretKeys(passphrase string) error {
	if e.Kind == EmailProviderKindSES && e.SES != nil && e.SES.SecretKey != "" {
		if err := e.SES.EncryptSecretKey(passphrase); err != nil {
			return err
		}
		e.SES.SecretKey = ""
	}

	if e.Kind == EmailProviderKindSMTP && e.SMTP != nil && e.SMTP.Password != "" {
		if err := e.SMTP.EncryptPassword(passphrase); err != nil {
			return err
		}
		e.SMTP.Password = ""
	}

	if e.Kind == EmailProviderKindSparkPost && e.SparkPost != nil && e.SparkPost.APIKey != "" {
		if err := e.SparkPost.EncryptAPIKey(passphrase); err != nil {
			return err
		}
		e.SparkPost.APIKey = ""
	}

	return nil
}

// DecryptSecretKeys decrypts all encrypted secret keys in the email provider
func (e *EmailProvider) DecryptSecretKeys(passphrase string) error {
	if e.Kind == EmailProviderKindSES && e.SES != nil && e.SES.EncryptedSecretKey != "" {
		if err := e.SES.DecryptSecretKey(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindSMTP && e.SMTP != nil && e.SMTP.EncryptedPassword != "" {
		if err := e.SMTP.DecryptPassword(passphrase); err != nil {
			return err
		}
	}

	if e.Kind == EmailProviderKindSparkPost && e.SparkPost != nil && e.SparkPost.EncryptedAPIKey != "" {
		if err := e.SparkPost.DecryptAPIKey(passphrase); err != nil {
			return err
		}
	}

	return nil
}

// SMTPSettings contains configuration for SMTP email server
type SMTPSettings struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"encrypted_password,omitempty"`
	SenderEmail       string `json:"sender_email"`
	UseTLS            bool   `json:"use_tls"`

	// decoded password, not stored in the database
	Password string `json:"password,omitempty"`
}

func (s *SMTPSettings) DecryptPassword(passphrase string) error {
	password, err := crypto.DecryptFromHexString(s.EncryptedPassword, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SMTP password: %w", err)
	}
	s.Password = password
	return nil
}

func (s *SMTPSettings) EncryptPassword(passphrase string) error {
	encryptedPassword, err := crypto.EncryptString(s.Password, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SMTP password: %w", err)
	}
	s.EncryptedPassword = encryptedPassword
	return nil
}

func (s *SMTPSettings) Validate(passphrase string) error {
	if s.Host == "" {
		return fmt.Errorf("host is required for SMTP configuration")
	}

	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("invalid port number for SMTP configuration: %d", s.Port)
	}

	if s.Username == "" {
		return fmt.Errorf("username is required for SMTP configuration")
	}

	if s.SenderEmail == "" {
		return fmt.Errorf("sender email is required for SMTP configuration")
	}

	if !govalidator.IsEmail(s.SenderEmail) {
		return fmt.Errorf("invalid sender email for SMTP configuration: %s", s.SenderEmail)
	}

	// Only encrypt password if it's not empty
	if s.Password != "" {
		if err := s.EncryptPassword(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SMTP password: %w", err)
		}
	}

	return nil
}

// SparkPostSettings contains configuration for SparkPost email provider
type SparkPostSettings struct {
	EncryptedAPIKey string `json:"encrypted_api_key,omitempty"`
	SenderEmail     string `json:"sender_email"`
	SandboxMode     bool   `json:"sandbox_mode"`

	// decoded API key, not stored in the database
	APIKey string `json:"api_key,omitempty"`
}

func (s *SparkPostSettings) DecryptAPIKey(passphrase string) error {
	apiKey, err := crypto.DecryptFromHexString(s.EncryptedAPIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to decrypt SparkPost API key: %w", err)
	}
	s.APIKey = apiKey
	return nil
}

func (s *SparkPostSettings) EncryptAPIKey(passphrase string) error {
	encryptedAPIKey, err := crypto.EncryptString(s.APIKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt SparkPost API key: %w", err)
	}
	s.EncryptedAPIKey = encryptedAPIKey
	return nil
}

func (s *SparkPostSettings) Validate(passphrase string) error {
	if s.SenderEmail == "" {
		return fmt.Errorf("sender email is required for SparkPost configuration")
	}

	if !govalidator.IsEmail(s.SenderEmail) {
		return fmt.Errorf("invalid sender email for SparkPost configuration: %s", s.SenderEmail)
	}

	// Encrypt API key if it's not empty
	if s.APIKey != "" {
		if err := s.EncryptAPIKey(passphrase); err != nil {
			return fmt.Errorf("failed to encrypt SparkPost API key: %w", err)
		}
	}

	return nil
}

// For database scanning
type dbWorkspace struct {
	ID        string
	Name      string
	Settings  []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ScanWorkspace scans a workspace from the database
func ScanWorkspace(scanner interface {
	Scan(dest ...interface{}) error
}) (*Workspace, error) {
	var dbw dbWorkspace
	if err := scanner.Scan(
		&dbw.ID,
		&dbw.Name,
		&dbw.Settings,
		&dbw.CreatedAt,
		&dbw.UpdatedAt,
	); err != nil {
		return nil, err
	}

	w := &Workspace{
		ID:        dbw.ID,
		Name:      dbw.Name,
		CreatedAt: dbw.CreatedAt,
		UpdatedAt: dbw.UpdatedAt,
	}

	if err := json.Unmarshal(dbw.Settings, &w.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return w, nil
}

// UserWorkspace represents the relationship between a user and a workspace
type UserWorkspace struct {
	UserID      string    `json:"user_id" db:"user_id"`
	WorkspaceID string    `json:"workspace_id" db:"workspace_id"`
	Role        string    `json:"role" db:"role"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// UserWorkspaceWithEmail extends UserWorkspace to include user email
type UserWorkspaceWithEmail struct {
	UserWorkspace
	Email string `json:"email" db:"email"`
}

// Validate performs validation on the user workspace fields
func (uw *UserWorkspace) Validate() error {
	if uw.UserID == "" {
		return fmt.Errorf("invalid user workspace: user_id is required")
	}
	if uw.WorkspaceID == "" {
		return fmt.Errorf("invalid user workspace: workspace_id is required")
	}
	if uw.Role == "" {
		return fmt.Errorf("invalid user workspace: role is required")
	}
	if uw.Role != "owner" && uw.Role != "member" {
		return fmt.Errorf("invalid user workspace: role must be either 'owner' or 'member'")
	}

	return nil
}

// WorkspaceInvitation represents an invitation to a workspace
type WorkspaceInvitation struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	InviterID   string    `json:"inviter_id"`
	Email       string    `json:"email"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id string) (*Workspace, error)
	List(ctx context.Context) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id string) error

	// User workspace management
	AddUserToWorkspace(ctx context.Context, userWorkspace *UserWorkspace) error
	RemoveUserFromWorkspace(ctx context.Context, userID string, workspaceID string) error
	GetUserWorkspaces(ctx context.Context, userID string) ([]*UserWorkspace, error)
	GetWorkspaceUsersWithEmail(ctx context.Context, workspaceID string) ([]*UserWorkspaceWithEmail, error)
	GetUserWorkspace(ctx context.Context, userID string, workspaceID string) (*UserWorkspace, error)

	// Workspace invitation management
	CreateInvitation(ctx context.Context, invitation *WorkspaceInvitation) error
	GetInvitationByID(ctx context.Context, id string) (*WorkspaceInvitation, error)
	GetInvitationByEmail(ctx context.Context, workspaceID, email string) (*WorkspaceInvitation, error)
	IsUserWorkspaceMember(ctx context.Context, userID, workspaceID string) (bool, error)

	// Database management
	GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error)
	CreateDatabase(ctx context.Context, workspaceID string) error
	DeleteDatabase(ctx context.Context, workspaceID string) error
}

// ErrUnauthorized is returned when a user is not authorized to perform an action
type ErrUnauthorized struct {
	Message string
}

func (e *ErrUnauthorized) Error() string {
	return e.Message
}

// WorkspaceServiceInterface defines the interface for workspace operations
type WorkspaceServiceInterface interface {
	CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string, fileManager FileManagerSettings) (*Workspace, error)
	GetWorkspace(ctx context.Context, id string) (*Workspace, error)
	ListWorkspaces(ctx context.Context) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone string, fileManager FileManagerSettings) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	GetWorkspaceMembersWithEmail(ctx context.Context, id string) ([]*UserWorkspaceWithEmail, error)
	InviteMember(ctx context.Context, workspaceID, email string) (*WorkspaceInvitation, string, error)
	AddUserToWorkspace(ctx context.Context, workspaceID string, userID string, role string) error
	RemoveUserFromWorkspace(ctx context.Context, workspaceID string, userID string) error
	TransferOwnership(ctx context.Context, workspaceID string, newOwnerID string, currentOwnerID string) error
}

// Request/Response types
type CreateWorkspaceRequest struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Settings WorkspaceSettings `json:"settings"`
}

func (r *CreateWorkspaceRequest) Validate(passphrase string) error {
	// Validate ID
	if r.ID == "" {
		return fmt.Errorf("invalid create workspace request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return fmt.Errorf("invalid create workspace request: id must be alphanumeric")
	}
	if len(r.ID) > 32 {
		return fmt.Errorf("invalid create workspace request: id length must be between 1 and 32")
	}

	// Validate Name
	if r.Name == "" {
		return fmt.Errorf("invalid create workspace request: name is required")
	}
	if len(r.Name) > 32 {
		return fmt.Errorf("invalid create workspace request: name length must be between 1 and 32")
	}

	// Validate Settings
	if err := r.Settings.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid create workspace request: %w", err)
	}

	return nil
}

type GetWorkspaceRequest struct {
	ID string `json:"id"`
}

type UpdateWorkspaceRequest struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Settings WorkspaceSettings `json:"settings"`
}

func (r *UpdateWorkspaceRequest) Validate(passphrase string) error {
	// Validate ID
	if r.ID == "" {
		return fmt.Errorf("invalid update workspace request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return fmt.Errorf("invalid update workspace request: id must be alphanumeric")
	}
	if len(r.ID) > 32 {
		return fmt.Errorf("invalid update workspace request: id length must be between 1 and 32")
	}

	// Validate Name
	if r.Name == "" {
		return fmt.Errorf("invalid update workspace request: name is required")
	}
	if len(r.Name) > 32 {
		return fmt.Errorf("invalid update workspace request: name length must be between 1 and 32")
	}

	// Validate Settings
	if err := r.Settings.Validate(passphrase); err != nil {
		return fmt.Errorf("invalid update workspace request: %w", err)
	}

	return nil
}

type DeleteWorkspaceRequest struct {
	ID string `json:"id"`
}

func (r *DeleteWorkspaceRequest) Validate() error {
	if r.ID == "" {
		return fmt.Errorf("invalid delete workspace request: id is required")
	}
	if !govalidator.IsAlphanumeric(r.ID) {
		return fmt.Errorf("invalid delete workspace request: id must be alphanumeric")
	}
	if len(r.ID) > 32 {
		return fmt.Errorf("invalid delete workspace request: id length must be between 1 and 32")
	}

	return nil
}

type InviteMemberRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Email       string `json:"email"`
}

func (r *InviteMemberRequest) Validate() error {
	if r.WorkspaceID == "" {
		return fmt.Errorf("invalid invite member request: workspace_id is required")
	}
	if !govalidator.IsAlphanumeric(r.WorkspaceID) {
		return fmt.Errorf("invalid invite member request: workspace_id must be alphanumeric")
	}
	if len(r.WorkspaceID) > 32 {
		return fmt.Errorf("invalid invite member request: workspace_id length must be between 1 and 32")
	}

	if r.Email == "" {
		return fmt.Errorf("invalid invite member request: email is required")
	}
	if !govalidator.IsEmail(r.Email) {
		return fmt.Errorf("invalid invite member request: email is not valid")
	}

	return nil
}
