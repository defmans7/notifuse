package domain

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
)

type Workspace struct {
	ID         string    `json:"id" valid:"required,alphanum,stringlength(1|20)"`
	Name       string    `json:"name" valid:"required,stringlength(1|255)"`
	WebsiteURL string    `json:"website_url" valid:"url,optional"`
	LogoURL    string    `json:"logo_url" valid:"url,optional"`
	Timezone   string    `json:"timezone" valid:"required,timezone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Validate performs validation on the workspace fields
func (w *Workspace) Validate() error {
	// Register custom validators
	govalidator.TagMap["timezone"] = govalidator.Validator(func(str string) bool {
		return IsValidTimezone(str)
	})

	_, err := govalidator.ValidateStruct(w)
	if err != nil {
		return fmt.Errorf("invalid workspace: %w", err)
	}

	return nil
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
