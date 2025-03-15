package domain

import (
	"context"
	"database/sql"
	"time"
)

type Workspace struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	WebsiteURL string    `json:"website_url"`
	LogoURL    string    `json:"logo_url"`
	Timezone   string    `json:"timezone"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
