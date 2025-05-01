package broadcast

import (
	"context"
	"errors"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	"github.com/Notifuse/notifuse/pkg/mjml"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestTemplateLoader_LoadTemplatesForBroadcast(t *testing.T) {
	// Setup test cases
	testCases := []struct {
		name            string
		broadcastID     string
		workspaceID     string
		setupMocks      func(*mocks.MockBroadcastSender, *mocks.MockTemplateService, *mocks.MockLogger)
		expectedError   bool
		expectedErrCode ErrorCode
		expectedCount   int
	}{
		{
			name:        "Success case with multiple templates",
			broadcastID: "broadcast-123",
			workspaceID: "workspace-1",
			setupMocks: func(mb *mocks.MockBroadcastSender, mt *mocks.MockTemplateService, ml *mocks.MockLogger) {
				// Setup broadcast mock
				broadcast := &domain.Broadcast{
					ID: "broadcast-123",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{ID: "var-1", TemplateID: "template-1"},
							{ID: "var-2", TemplateID: "template-2"},
						},
					},
				}

				mb.EXPECT().
					GetBroadcast(gomock.Any(), "workspace-1", "broadcast-123").
					Return(broadcast, nil)

				// Setup template mocks
				template1 := &domain.Template{ID: "template-1", Name: "Template 1"}
				template2 := &domain.Template{ID: "template-2", Name: "Template 2"}

				mt.EXPECT().
					GetTemplateByID(gomock.Any(), "workspace-1", "template-1", int64(1)).
					Return(template1, nil)

				mt.EXPECT().
					GetTemplateByID(gomock.Any(), "workspace-1", "template-2", int64(1)).
					Return(template2, nil)

				// Setup logger expectations
				ml.EXPECT().WithFields(gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Debug(gomock.Any()).AnyTimes()
				ml.EXPECT().Info(gomock.Any()).AnyTimes()
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:        "Broadcast not found",
			broadcastID: "broadcast-not-found",
			workspaceID: "workspace-1",
			setupMocks: func(mb *mocks.MockBroadcastSender, mt *mocks.MockTemplateService, ml *mocks.MockLogger) {
				// Setup broadcast mock to return error
				mb.EXPECT().
					GetBroadcast(gomock.Any(), "workspace-1", "broadcast-not-found").
					Return(nil, errors.New("broadcast not found"))

				// Setup logger expectations
				ml.EXPECT().WithFields(gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Debug(gomock.Any()).AnyTimes()
				ml.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			expectedError:   true,
			expectedErrCode: ErrCodeBroadcastNotFound,
		},
		{
			name:        "No variations in broadcast",
			broadcastID: "broadcast-no-variations",
			workspaceID: "workspace-1",
			setupMocks: func(mb *mocks.MockBroadcastSender, mt *mocks.MockTemplateService, ml *mocks.MockLogger) {
				// Setup broadcast mock with no variations
				broadcast := &domain.Broadcast{
					ID: "broadcast-no-variations",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{}, // Empty variations
					},
				}

				mb.EXPECT().
					GetBroadcast(gomock.Any(), "workspace-1", "broadcast-no-variations").
					Return(broadcast, nil)

				// Setup logger expectations
				ml.EXPECT().WithFields(gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Debug(gomock.Any()).AnyTimes()
				ml.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			expectedError:   true,
			expectedErrCode: ErrCodeTemplateMissing,
		},
		{
			name:        "Template load failure",
			broadcastID: "broadcast-template-failure",
			workspaceID: "workspace-1",
			setupMocks: func(mb *mocks.MockBroadcastSender, mt *mocks.MockTemplateService, ml *mocks.MockLogger) {
				// Setup broadcast mock
				broadcast := &domain.Broadcast{
					ID: "broadcast-template-failure",
					TestSettings: domain.BroadcastTestSettings{
						Variations: []domain.BroadcastVariation{
							{ID: "var-1", TemplateID: "template-error"},
						},
					},
				}

				mb.EXPECT().
					GetBroadcast(gomock.Any(), "workspace-1", "broadcast-template-failure").
					Return(broadcast, nil)

				// Setup template mock to fail
				mt.EXPECT().
					GetTemplateByID(gomock.Any(), "workspace-1", "template-error", int64(1)).
					Return(nil, errors.New("template not found"))

				// Setup logger expectations
				ml.EXPECT().WithFields(gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Debug(gomock.Any()).AnyTimes()
				ml.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			expectedError:   true,
			expectedErrCode: ErrCodeTemplateMissing,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockBroadcastSender := mocks.NewMockBroadcastSender(ctrl)
			mockTemplateService := mocks.NewMockTemplateService(ctrl)
			mockLogger := mocks.NewMockLogger(ctrl)

			// Configure mocks
			tc.setupMocks(mockBroadcastSender, mockTemplateService, mockLogger)

			// Create template loader with mocks
			loader := NewTemplateLoader(
				mockBroadcastSender,
				mockTemplateService,
				mockLogger,
				nil, // Use default config
			)

			// Execute the function
			templates, err := loader.LoadTemplatesForBroadcast(
				context.Background(),
				tc.workspaceID,
				tc.broadcastID,
			)

			// Assert error state
			if tc.expectedError {
				assert.Error(t, err)
				// Check the error type
				if err != nil {
					var broadcastErr *BroadcastError
					assert.True(t, errors.As(err, &broadcastErr))
					if broadcastErr != nil {
						assert.Equal(t, tc.expectedErrCode, broadcastErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, templates)
				assert.Equal(t, tc.expectedCount, len(templates))
			}
		})
	}
}

func TestTemplateLoader_ValidateTemplates(t *testing.T) {
	// Setup test cases
	testCases := []struct {
		name          string
		templates     map[string]*domain.Template
		setupMocks    func(*mocks.MockLogger)
		expectedError bool
		expectedCode  ErrorCode
	}{
		{
			name: "Valid templates",
			templates: map[string]*domain.Template{
				"template-1": {
					ID:   "template-1",
					Name: "Valid Template",
					Email: &domain.EmailTemplate{
						FromAddress: "test@example.com",
						FromName:    "Test Sender",
						Subject:     "Test Subject",
						VisualEditorTree: mjml.EmailBlock{
							Kind: "root",
							Data: map[string]interface{}{
								"styles": map[string]interface{}{},
							},
						},
					},
				},
			},
			setupMocks: func(ml *mocks.MockLogger) {
				ml.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(ml).AnyTimes()
			},
			expectedError: false,
		},
		{
			name:      "Empty templates",
			templates: map[string]*domain.Template{},
			setupMocks: func(ml *mocks.MockLogger) {
				ml.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(ml).AnyTimes()
			},
			expectedError: true,
			expectedCode:  ErrCodeTemplateMissing,
		},
		{
			name: "Nil template",
			templates: map[string]*domain.Template{
				"template-nil": nil,
			},
			setupMocks: func(ml *mocks.MockLogger) {
				ml.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(ml).AnyTimes()
			},
			expectedError: true,
			expectedCode:  ErrCodeTemplateInvalid,
		},
		{
			name: "Missing email config",
			templates: map[string]*domain.Template{
				"template-no-email": {
					ID:   "template-no-email",
					Name: "Template Without Email",
					// No Email field
				},
			},
			setupMocks: func(ml *mocks.MockLogger) {
				ml.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			expectedError: true,
			expectedCode:  ErrCodeTemplateInvalid,
		},
		{
			name: "Missing from address",
			templates: map[string]*domain.Template{
				"template-no-from": {
					ID:   "template-no-from",
					Name: "Template Without From",
					Email: &domain.EmailTemplate{
						// FromAddress missing
						FromName: "Test Sender",
						Subject:  "Test Subject",
						VisualEditorTree: mjml.EmailBlock{
							Kind: "root",
							Data: map[string]interface{}{
								"styles": map[string]interface{}{},
							},
						},
					},
				},
			},
			setupMocks: func(ml *mocks.MockLogger) {
				ml.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(ml).AnyTimes()
				ml.EXPECT().Error(gomock.Any()).AnyTimes()
			},
			expectedError: true,
			expectedCode:  ErrCodeTemplateInvalid,
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := mocks.NewMockLogger(ctrl)

			// Configure mocks
			tc.setupMocks(mockLogger)

			// Create loader with only the logger
			loader := &templateLoader{
				logger: mockLogger,
				config: DefaultConfig(),
			}

			// Execute validation
			err := loader.ValidateTemplates(tc.templates)

			// Assert error state
			if tc.expectedError {
				assert.Error(t, err)
				// Check error type
				if err != nil {
					var broadcastErr *BroadcastError
					assert.True(t, errors.As(err, &broadcastErr))
					if broadcastErr != nil {
						assert.Equal(t, tc.expectedCode, broadcastErr.Code)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
