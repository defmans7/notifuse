# Testing Challenges and Recommendations for TransactionalNotificationService

## Overview

The TransactionalNotificationService provides operations for managing and sending transactional notifications, with key methods `SendNotification` and `DoSendEmailNotification` that remain untested. This report outlines the challenges encountered in testing these methods and provides actionable recommendations to improve testability.

## Current Testing Status

- Test coverage for CRUD operations (Create, Update, Get, List, Delete): ~100%
- Test coverage for SendNotification: 0%
- Test coverage for DoSendEmailNotification: 0%

Tests for both methods are currently skipped with comments indicating "complex setup with email service and templates" and "complex setup with template compilation" respectively.

## Key Testing Challenges

### 1. Complex Dependencies

The `TransactionalNotificationService` has multiple complex dependencies:

```go
type TransactionalNotificationService struct {
    transactionalRepo  domain.TransactionalNotificationRepository
    messageHistoryRepo domain.MessageHistoryRepository
    templateService    domain.TemplateService
    contactService     domain.ContactService
    emailService       *EmailService  // Concrete type, not an interface
    logger             logger.Logger
}
```

Specifically, the `emailService` is a concrete type (`*EmailService`) rather than an interface, making it difficult to mock in tests.

### 2. Direct Email Provider Dependencies

The `DoSendEmailNotification` method interacts with various email provider APIs (SES, Mailjet, Mailgun, etc.) which complicates testing by requiring external service mocking.

### 3. Complex Template Compilation Process

Template compilation involves converting MJML templates to HTML with data substitution, which is difficult to mock and requires complex test fixtures.

### 4. Complex Parameter Structures

Both methods have numerous complex parameters and nested structures, making it challenging to create representative test cases.

### 5. Concrete Type EmailService vs Interface

A fundamental issue is that `emailService` is a concrete type pointer (`*EmailService`) rather than an interface. This tightly couples the implementation and makes mocking difficult.

## Recommendations

### 1. Create EmailService Interface

Introduce an interface for EmailService in the domain package:

```go
// EmailServiceInterface defines methods that email service should implement
type EmailServiceInterface interface {
    SendEmail(ctx context.Context, workspaceID string, isMarketing bool,
              fromAddress string, fromName string, to string,
              subject string, content string,
              optionalProvider ...*EmailProvider) error
    TestEmailProvider(ctx context.Context, workspaceID string, provider EmailProvider, to string) error
    TestTemplate(ctx context.Context, workspaceID string, templateID string, integrationID string, recipientEmail string) error
    SetHTTPClient(client HTTPClient)
}
```

### 2. Update TransactionalNotificationService to Use Interface

```go
type TransactionalNotificationService struct {
    transactionalRepo  domain.TransactionalNotificationRepository
    messageHistoryRepo domain.MessageHistoryRepository
    templateService    domain.TemplateService
    contactService     domain.ContactService
    emailService       domain.EmailServiceInterface  // Interface instead of concrete type
    logger             logger.Logger
}
```

### 3. Create Test Fixtures for Email Templates

Develop a set of pre-compiled email templates for testing:

```go
var testEmailTemplate = &domain.Template{
    ID:      "template-test-id",
    Name:    "Test Template",
    Version: 1,
    Email: &domain.EmailTemplate{
        Subject:     "Test Subject",
        FromName:    "Test Sender",
        FromAddress: "test@example.com",
        VisualEditorTree: mjml.EmailBlock{
            Kind: "root",
            Data: map[string]interface{}{
                "styles": map[string]interface{}{},
            },
        },
    },
}
```

### 4. Create Helper Functions for Domain Objects

Create helper functions to easily build test objects:

```go
func createTestContact() *domain.Contact {
    return &domain.Contact{
        Email: "test@example.com",
        FirstName: &domain.NullableString{
            Valid:  true,
            String: "Test",
        },
        LastName: &domain.NullableString{
            Valid:  true,
            String: "User",
        },
    }
}
```

### 5. Implement Comprehensive Tests

Create tests for multiple scenarios:

#### For SendNotification:

- Successful notification send through email channel
- Notification not found
- Contact validation failure
- Multiple channels
- Template compilation error

#### For DoSendEmailNotification:

- Successful email sending
- Template not found
- Template compilation error
- Email service error

### 6. Use a Testable HTTP Client

For testing email provider API calls, implement a mock HTTP client:

```go
type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    return m.DoFunc(req)
}
```

### 7. Extract Email Provider Logic

Consider extracting email provider-specific logic to separate services that can be independently tested.

## Implementation Examples

### SendNotification Test Example

```go
func TestTransactionalNotificationService_SendNotification(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockTransactionalNotificationRepository(ctrl)
    mockMsgHistoryRepo := mocks.NewMockMessageHistoryRepository(ctrl)
    mockTemplateService := mocks.NewMockTemplateService(ctrl)
    mockContactService := mocks.NewMockContactService(ctrl)
    mockEmailService := mocks.NewMockEmailServiceInterface(ctrl)
    mockLogger := pkgmocks.NewMockLogger(ctrl)

    // Logger setup...
    mockLogger.EXPECT().WithFields(gomock.Any()).Return(mockLogger).AnyTimes()
    mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
    mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
    mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()

    ctx := context.Background()
    workspace := "test-workspace"
    notificationID := "test-notification"
    templateID := "test-template"

    // Create test notification
    notification := &domain.TransactionalNotification{
        ID: notificationID,
        Channels: map[domain.TransactionalChannel]domain.ChannelTemplate{
            domain.TransactionalChannelEmail: {
                TemplateID: templateID,
            },
        },
    }

    // Create test contact
    contact := createTestContact()

    // Test cases...
    t.Run("Success_SendEmailNotification", func(t *testing.T) {
        // Expect repo to get notification
        mockRepo.EXPECT().
            Get(gomock.Any(), workspace, notificationID).
            Return(notification, nil)

        // Contact upsert operation
        upsertOp := domain.UpsertContactOperation{
            Action: domain.UpsertContactOperationUpdate,
            Contact: contact,
        }
        mockContactService.EXPECT().
            UpsertContact(gomock.Any(), workspace, gomock.Any()).
            Return(upsertOp)

        // Expect contact retrieval
        mockContactService.EXPECT().
            GetContactByEmail(gomock.Any(), workspace, contact.Email).
            Return(contact, nil)

        // Expect successful email sending call
        mockEmailService.EXPECT().
            SendEmail(gomock.Any(), gomock.Any(), gomock.Any(),
                     gomock.Any(), gomock.Any(), gomock.Any(),
                     gomock.Any(), gomock.Any(), gomock.Any()).
            Return(nil)

        // Create the service
        service := &TransactionalNotificationService{
            transactionalRepo:  mockRepo,
            messageHistoryRepo: mockMsgHistoryRepo,
            templateService:    mockTemplateService,
            contactService:     mockContactService,
            emailService:       mockEmailService,
            logger:             mockLogger,
        }

        // Call the method
        sendParams := domain.TransactionalNotificationSendParams{
            ID:      notificationID,
            Contact: contact,
        }

        messageID, err := service.SendNotification(ctx, workspace, sendParams)

        // Assert results
        assert.NoError(t, err)
        assert.NotEmpty(t, messageID)
    })
}
```

## Conclusion

The key to making these methods testable is decoupling them from concrete implementations through interfaces and creating appropriate mocks. The most important change is converting the EmailService dependency from a concrete type to an interface.

By implementing these recommendations, the SendNotification and DoSendEmailNotification methods can be comprehensively tested, improving overall code quality and reliability.
