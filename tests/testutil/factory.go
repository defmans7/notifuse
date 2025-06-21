package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/notifuse_mjml"
	"github.com/google/uuid"
)

// TestDataFactory creates test data entities using domain repositories
type TestDataFactory struct {
	db                 *sql.DB
	userRepo           domain.UserRepository
	workspaceRepo      domain.WorkspaceRepository
	contactRepo        domain.ContactRepository
	listRepo           domain.ListRepository
	templateRepo       domain.TemplateRepository
	broadcastRepo      domain.BroadcastRepository
	messageHistoryRepo domain.MessageHistoryRepository
	contactListRepo    domain.ContactListRepository
}

// NewTestDataFactory creates a new test data factory with repository dependencies
func NewTestDataFactory(
	db *sql.DB,
	userRepo domain.UserRepository,
	workspaceRepo domain.WorkspaceRepository,
	contactRepo domain.ContactRepository,
	listRepo domain.ListRepository,
	templateRepo domain.TemplateRepository,
	broadcastRepo domain.BroadcastRepository,
	messageHistoryRepo domain.MessageHistoryRepository,
	contactListRepo domain.ContactListRepository,
) *TestDataFactory {
	return &TestDataFactory{
		db:                 db,
		userRepo:           userRepo,
		workspaceRepo:      workspaceRepo,
		contactRepo:        contactRepo,
		listRepo:           listRepo,
		templateRepo:       templateRepo,
		broadcastRepo:      broadcastRepo,
		messageHistoryRepo: messageHistoryRepo,
		contactListRepo:    contactListRepo,
	}
}

// CreateUser creates a test user using the user repository
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

	err := tdf.userRepo.CreateUser(context.Background(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user via repository: %w", err)
	}

	return user, nil
}

// CreateWorkspace creates a test workspace using the workspace repository
func (tdf *TestDataFactory) CreateWorkspace(opts ...WorkspaceOption) (*domain.Workspace, error) {
	workspace := &domain.Workspace{
		ID:   fmt.Sprintf("test%s", uuid.New().String()[:8]), // Keep it under 20 chars
		Name: fmt.Sprintf("Test Workspace %s", uuid.New().String()[:8]),
		Settings: domain.WorkspaceSettings{
			Timezone:  "UTC",
			SecretKey: fmt.Sprintf("test-secret-key-%s", uuid.New().String()[:8]),
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(workspace)
	}

	err := tdf.workspaceRepo.Create(context.Background(), workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace via repository: %w", err)
	}

	return workspace, nil
}

// CreateContact creates a test contact using the contact repository
func (tdf *TestDataFactory) CreateContact(workspaceID string, opts ...ContactOption) (*domain.Contact, error) {
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

	// Use UpsertContact since that's the method available in the repository
	_, err := tdf.contactRepo.UpsertContact(context.Background(), workspaceID, contact)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact via repository: %w", err)
	}

	return contact, nil
}

// CreateList creates a test list using the list repository
func (tdf *TestDataFactory) CreateList(workspaceID string, opts ...ListOption) (*domain.List, error) {
	list := &domain.List{
		ID:            fmt.Sprintf("list%s", uuid.New().String()[:8]), // Keep it under 32 chars
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

	err := tdf.listRepo.CreateList(context.Background(), workspaceID, list)
	if err != nil {
		return nil, fmt.Errorf("failed to create list via repository: %w", err)
	}

	return list, nil
}

// CreateTemplate creates a test template using the template repository
func (tdf *TestDataFactory) CreateTemplate(workspaceID string, opts ...TemplateOption) (*domain.Template, error) {
	template := &domain.Template{
		ID:        fmt.Sprintf("tmpl%s", uuid.New().String()[:8]), // Keep it under 32 chars
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

	err := tdf.templateRepo.CreateTemplate(context.Background(), workspaceID, template)
	if err != nil {
		return nil, fmt.Errorf("failed to create template via repository: %w", err)
	}

	return template, nil
}

// CreateBroadcast creates a test broadcast using the broadcast repository
func (tdf *TestDataFactory) CreateBroadcast(workspaceID string, opts ...BroadcastOption) (*domain.Broadcast, error) {
	broadcast := &domain.Broadcast{
		ID:           fmt.Sprintf("bc%s", uuid.New().String()[:8]), // Keep it under 32 chars
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

	err := tdf.broadcastRepo.CreateBroadcast(context.Background(), broadcast)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast via repository: %w", err)
	}

	return broadcast, nil
}

// CreateMessageHistory creates a test message history using the message history repository
func (tdf *TestDataFactory) CreateMessageHistory(workspaceID string, opts ...MessageHistoryOption) (*domain.MessageHistory, error) {
	now := time.Now().UTC()
	message := &domain.MessageHistory{
		ID:              uuid.New().String(),
		ContactEmail:    fmt.Sprintf("contact-%s@example.com", uuid.New().String()[:8]),
		TemplateID:      uuid.New().String(),
		TemplateVersion: 1,
		Channel:         "email",
		MessageData: domain.MessageData{
			Data: map[string]interface{}{
				"subject": "Test Message",
				"body":    "This is a test message",
			},
			Metadata: map[string]interface{}{
				"test": true,
			},
		},
		SentAt:    now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Apply options
	for _, opt := range opts {
		opt(message)
	}

	err := tdf.messageHistoryRepo.Create(context.Background(), workspaceID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to create message history via repository: %w", err)
	}

	return message, nil
}

// CreateContactList creates a test contact list relationship using the repository
func (tdf *TestDataFactory) CreateContactList(workspaceID string, opts ...ContactListOption) (*domain.ContactList, error) {
	contactList := &domain.ContactList{
		Email:     fmt.Sprintf("contact-%s@example.com", uuid.New().String()[:8]),
		ListID:    uuid.New().String(),
		Status:    domain.ContactListStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Apply options
	for _, opt := range opts {
		opt(contactList)
	}

	err := tdf.contactListRepo.AddContactToList(context.Background(), workspaceID, contactList)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact list via repository: %w", err)
	}

	return contactList, nil
}

// AddUserToWorkspace adds a user to a workspace with the specified role
func (tdf *TestDataFactory) AddUserToWorkspace(userID, workspaceID, role string) error {
	userWorkspace := &domain.UserWorkspace{
		UserID:      userID,
		WorkspaceID: workspaceID,
		Role:        role,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	err := tdf.workspaceRepo.AddUserToWorkspace(context.Background(), userWorkspace)
	if err != nil {
		return fmt.Errorf("failed to add user to workspace: %w", err)
	}

	return nil
}

// Option types for customizing test data
type UserOption func(*domain.User)
type WorkspaceOption func(*domain.Workspace)
type ContactOption func(*domain.Contact)
type ListOption func(*domain.List)
type TemplateOption func(*domain.Template)
type BroadcastOption func(*domain.Broadcast)
type MessageHistoryOption func(*domain.MessageHistory)
type ContactListOption func(*domain.ContactList)

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

func WithBroadcastABTesting(templateIDs []string) BroadcastOption {
	return func(b *domain.Broadcast) {
		if len(templateIDs) >= 2 {
			b.TestSettings.Enabled = true
			b.TestSettings.SamplePercentage = 50
			b.TestSettings.AutoSendWinner = true
			b.TestSettings.AutoSendWinnerMetric = "open_rate"
			b.TestSettings.TestDurationHours = 24
			// Create variations for the templates
			variations := make([]domain.BroadcastVariation, len(templateIDs))
			for i, templateID := range templateIDs {
				variations[i] = domain.BroadcastVariation{
					VariationName: fmt.Sprintf("Version %c", 'A'+i),
					TemplateID:    templateID,
				}
			}
			b.TestSettings.Variations = variations
		}
	}
}

// Message history options
func WithMessageHistoryContactEmail(email string) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.ContactEmail = email
	}
}

func WithMessageHistoryTemplateID(templateID string) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.TemplateID = templateID
	}
}

func WithMessageHistoryTemplateVersion(version int64) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.TemplateVersion = version
	}
}

func WithMessageHistoryChannel(channel string) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.Channel = channel
	}
}

// ContactList options
func WithContactListEmail(email string) ContactListOption {
	return func(cl *domain.ContactList) {
		cl.Email = email
	}
}

func WithContactListListID(listID string) ContactListOption {
	return func(cl *domain.ContactList) {
		cl.ListID = listID
	}
}

func WithContactListStatus(status domain.ContactListStatus) ContactListOption {
	return func(cl *domain.ContactList) {
		cl.Status = status
	}
}

// Convenience aliases for cleaner test code
func WithMessageContact(email string) MessageHistoryOption {
	return WithMessageHistoryContactEmail(email)
}

func WithMessageTemplate(templateID string) MessageHistoryOption {
	return WithMessageHistoryTemplateID(templateID)
}

func WithMessageChannel(channel string) MessageHistoryOption {
	return WithMessageHistoryChannel(channel)
}

func WithMessageBroadcast(broadcastID string) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.BroadcastID = &broadcastID
	}
}

func WithMessageSentAt(sentAt time.Time) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		m.SentAt = sentAt
	}
}

func WithMessageDelivered(delivered bool) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		if delivered {
			now := time.Now().UTC()
			m.DeliveredAt = &now
		} else {
			m.DeliveredAt = nil
		}
	}
}

func WithMessageOpened(opened bool) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		if opened {
			now := time.Now().UTC()
			m.OpenedAt = &now
			// If opened, also mark as delivered
			if m.DeliveredAt == nil {
				m.DeliveredAt = &now
			}
		} else {
			m.OpenedAt = nil
		}
	}
}

func WithMessageClicked(clicked bool) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		if clicked {
			now := time.Now().UTC()
			m.ClickedAt = &now
			// If clicked, also mark as opened and delivered
			if m.OpenedAt == nil {
				m.OpenedAt = &now
			}
			if m.DeliveredAt == nil {
				m.DeliveredAt = &now
			}
		} else {
			m.ClickedAt = nil
		}
	}
}

func WithMessageFailed(failed bool) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		if failed {
			now := time.Now().UTC()
			m.FailedAt = &now
		} else {
			m.FailedAt = nil
		}
	}
}

func WithMessageBounced(bounced bool) MessageHistoryOption {
	return func(m *domain.MessageHistory) {
		if bounced {
			now := time.Now().UTC()
			m.BouncedAt = &now
		} else {
			m.BouncedAt = nil
		}
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
		VisualEditorTree: createDefaultMJMLBlock(),
	}
}

func createDefaultMJMLBlock() notifuse_mjml.EmailBlock {
	// Create a simple text block for testing - avoid complex nested structures
	textBlock := notifuse_mjml.BaseBlock{
		ID:   "text-1",
		Type: notifuse_mjml.MJMLComponentMjText,
		Attributes: map[string]interface{}{
			"content": "Hello Test!",
		},
		Children: []interface{}{},
	}

	return &textBlock
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
