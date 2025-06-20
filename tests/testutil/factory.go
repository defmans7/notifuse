package testutil

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/google/uuid"
)

// TestDataFactory creates test data entities
type TestDataFactory struct {
	db *sql.DB
}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory(db *sql.DB) *TestDataFactory {
	return &TestDataFactory{db: db}
}

// CreateUser creates a test user
func (tdf *TestDataFactory) CreateUser(opts ...UserOption) (*domain.User, error) {
	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     fmt.Sprintf("user-%s@example.com", uuid.New().String()[:8]),
		Name:      "Test User",
		Type:      domain.UserTypeUser,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(user)
	}

	query := `
		INSERT INTO users (id, email, name, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tdf.db.Exec(query, user.ID, user.Email, user.Name, user.Type, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// CreateWorkspace creates a test workspace
func (tdf *TestDataFactory) CreateWorkspace(opts ...WorkspaceOption) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:        uuid.New().String(),
		Name:      fmt.Sprintf("Test Workspace %s", uuid.New().String()[:8]),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(workspace)
	}

	// Note: Using simplified query since Workspace has complex Settings and Integrations fields
	query := `
		INSERT INTO workspaces (id, name, settings, integrations, created_at, updated_at)
		VALUES ($1, $2, '{}', '[]', $3, $4)
	`
	_, err := tdf.db.Exec(query, workspace.ID, workspace.Name, workspace.CreatedAt, workspace.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return workspace, nil
}

// CreateContact creates a test contact
func (tdf *TestDataFactory) CreateContact(opts ...ContactOption) (*domain.Contact, error) {
	contact := &domain.Contact{
		Email:     fmt.Sprintf("contact-%s@example.com", uuid.New().String()[:8]),
		FirstName: &domain.NullableString{String: "Test", IsNull: false},
		LastName:  &domain.NullableString{String: "Contact", IsNull: false},
		Timezone:  &domain.NullableString{String: "UTC", IsNull: false},
		Language:  &domain.NullableString{String: "en", IsNull: false},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(contact)
	}

	query := `
		INSERT INTO contacts (email, external_id, timezone, language, first_name, last_name, 
			phone, address_line_1, address_line_2, country, postcode, state, job_title,
			lifetime_value, orders_count, last_order_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	_, err := tdf.db.Exec(query,
		contact.Email, contact.ExternalID, contact.Timezone, contact.Language,
		contact.FirstName, contact.LastName, contact.Phone, contact.AddressLine1,
		contact.AddressLine2, contact.Country, contact.Postcode, contact.State,
		contact.JobTitle, contact.LifetimeValue, contact.OrdersCount, contact.LastOrderAt,
		contact.CreatedAt, contact.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return contact, nil
}

// CreateList creates a test list
func (tdf *TestDataFactory) CreateList(opts ...ListOption) (*domain.List, error) {
	list := &domain.List{
		ID:            uuid.New().String(),
		Name:          fmt.Sprintf("Test List %s", uuid.New().String()[:8]),
		IsDoubleOptin: false,
		IsPublic:      false,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(list)
	}

	query := `
		INSERT INTO lists (id, name, is_double_optin, is_public, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tdf.db.Exec(query, list.ID, list.Name, list.IsDoubleOptin, list.IsPublic,
		list.Description, list.CreatedAt, list.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	return list, nil
}

// CreateTemplate creates a test template
func (tdf *TestDataFactory) CreateTemplate(opts ...TemplateOption) (*domain.Template, error) {
	template := &domain.Template{
		ID:        uuid.New().String(),
		Name:      fmt.Sprintf("Test Template %s", uuid.New().String()[:8]),
		Version:   1,
		Channel:   "email",
		Category:  "marketing",
		Email:     createDefaultEmailTemplate(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(template)
	}

	query := `
		INSERT INTO templates (id, name, version, channel, email, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := tdf.db.Exec(query, template.ID, template.Name, template.Version,
		template.Channel, template.Email, template.Category, template.CreatedAt, template.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// CreateBroadcast creates a test broadcast
func (tdf *TestDataFactory) CreateBroadcast(workspaceID string, opts ...BroadcastOption) (*domain.Broadcast, error) {
	broadcast := &domain.Broadcast{
		ID:           uuid.New().String(),
		WorkspaceID:  workspaceID,
		Name:         fmt.Sprintf("Test Broadcast %s", uuid.New().String()[:8]),
		Status:       domain.BroadcastStatusDraft,
		Audience:     createDefaultAudience(),
		Schedule:     createDefaultSchedule(),
		TestSettings: createDefaultTestSettings(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(broadcast)
	}

	query := `
		INSERT INTO broadcasts (id, workspace_id, name, status, audience, schedule, test_settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := tdf.db.Exec(query, broadcast.ID, broadcast.WorkspaceID, broadcast.Name,
		broadcast.Status, broadcast.Audience, broadcast.Schedule, broadcast.TestSettings,
		broadcast.CreatedAt, broadcast.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	return broadcast, nil
}

// Option types for customizing test data
type UserOption func(*domain.User)
type WorkspaceOption func(*domain.Workspace)
type ContactOption func(*domain.Contact)
type ListOption func(*domain.List)
type TemplateOption func(*domain.Template)
type BroadcastOption func(*domain.Broadcast)

// User options
func WithUserEmail(email string) UserOption {
	return func(u *domain.User) {
		u.Email = email
	}
}

func WithUserName(name string) UserOption {
	return func(u *domain.User) {
		u.Name = name
	}
}

func WithUserType(userType domain.UserType) UserOption {
	return func(u *domain.User) {
		u.Type = userType
	}
}

// Workspace options
func WithWorkspaceName(name string) WorkspaceOption {
	return func(w *domain.Workspace) {
		w.Name = name
	}
}

func WithWorkspaceSettings(settings domain.WorkspaceSettings) WorkspaceOption {
	return func(w *domain.Workspace) {
		w.Settings = settings
	}
}

// Contact options
func WithContactEmail(email string) ContactOption {
	return func(c *domain.Contact) {
		c.Email = email
	}
}

func WithContactName(firstName, lastName string) ContactOption {
	return func(c *domain.Contact) {
		c.FirstName = &domain.NullableString{String: firstName, IsNull: false}
		c.LastName = &domain.NullableString{String: lastName, IsNull: false}
	}
}

func WithContactExternalID(externalID string) ContactOption {
	return func(c *domain.Contact) {
		c.ExternalID = &domain.NullableString{String: externalID, IsNull: false}
	}
}

// List options
func WithListName(name string) ListOption {
	return func(l *domain.List) {
		l.Name = name
	}
}

func WithListDoubleOptin(enabled bool) ListOption {
	return func(l *domain.List) {
		l.IsDoubleOptin = enabled
	}
}

// Template options
func WithTemplateName(name string) TemplateOption {
	return func(t *domain.Template) {
		t.Name = name
	}
}

func WithTemplateCategory(category string) TemplateOption {
	return func(t *domain.Template) {
		t.Category = category
	}
}

// Broadcast options
func WithBroadcastName(name string) BroadcastOption {
	return func(b *domain.Broadcast) {
		b.Name = name
	}
}

func WithBroadcastStatus(status domain.BroadcastStatus) BroadcastOption {
	return func(b *domain.Broadcast) {
		b.Status = status
	}
}

// Helper functions to create default structures
func createDefaultEmailTemplate() *domain.EmailTemplate {
	return &domain.EmailTemplate{
		Subject: "Test Email Subject",
		CompiledPreview: `<mjml>
			<mj-body>
				<mj-section>
					<mj-column>
						<mj-text>Hello Test!</mj-text>
					</mj-column>
				</mj-section>
			</mj-body>
		</mjml>`,
	}
}

func createDefaultAudience() domain.AudienceSettings {
	return domain.AudienceSettings{
		ExcludeUnsubscribed: true,
		SkipDuplicateEmails: true,
	}
}

func createDefaultSchedule() domain.ScheduleSettings {
	return domain.ScheduleSettings{
		IsScheduled: false,
	}
}

func createDefaultTestSettings() domain.BroadcastTestSettings {
	return domain.BroadcastTestSettings{
		Enabled:          false,
		SamplePercentage: 100,
		Variations: []domain.BroadcastVariation{
			{
				VariationName: "Default",
				TemplateID:    "",
			},
		},
	}
}
