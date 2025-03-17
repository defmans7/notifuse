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

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id string) (*Workspace, error)
	List(ctx context.Context) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id string) error

	// Database management
	GetConnection(ctx context.Context, workspaceID string) (*sql.DB, error)
	CreateDatabase(ctx context.Context, workspaceID string) error
	DeleteDatabase(ctx context.Context, workspaceID string) error
}
