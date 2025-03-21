package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
)

// WorkspaceSettings contains configurable workspace settings
type WorkspaceSettings struct {
	WebsiteURL string `json:"website_url,omitempty" valid:"url,optional"`
	LogoURL    string `json:"logo_url,omitempty" valid:"url,optional"`
	CoverURL   string `json:"cover_url,omitempty" valid:"url,optional"`
	Timezone   string `json:"timezone" valid:"required,timezone"`
}

type Workspace struct {
	ID        string            `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name      string            `json:"name" valid:"required,stringlength(1|255)"`
	Settings  WorkspaceSettings `json:"settings"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// Validate performs validation on the workspace fields
func (w *Workspace) Validate() error {
	// Register custom validators
	govalidator.TagMap["timezone"] = govalidator.Validator(func(str string) bool {
		return IsValidTimezone(str)
	})

	// First validate the workspace itself
	if _, err := govalidator.ValidateStruct(w); err != nil {
		return fmt.Errorf("invalid workspace: %w", err)
	}

	// Then validate the settings
	if _, err := govalidator.ValidateStruct(&w.Settings); err != nil {
		return fmt.Errorf("invalid workspace settings: %w", err)
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
	Role        string    `json:"role" db:"role" valid:"required,in(owner|member)"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Validate performs validation on the user workspace fields
func (uw *UserWorkspace) Validate() error {
	if _, err := govalidator.ValidateStruct(uw); err != nil {
		return fmt.Errorf("invalid user workspace: %w", err)
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
	GetWorkspaceUsers(ctx context.Context, workspaceID string) ([]*UserWorkspace, error)
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
	CreateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone, ownerID string) (*Workspace, error)
	GetWorkspace(ctx context.Context, id, requesterID string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, userID string) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, id, name, websiteURL, logoURL, coverURL, timezone, ownerID string) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, id, requesterID string) error
	GetWorkspaceMembers(ctx context.Context, id, requesterID string) ([]*UserWorkspace, error)
	InviteMember(ctx context.Context, workspaceID, inviterID, email string) (*WorkspaceInvitation, string, error)
}
