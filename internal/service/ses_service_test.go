package service

import (
	"context"
	"testing"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/internal/domain/mocks"
	pkgmocks "github.com/Notifuse/notifuse/pkg/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Helper function to create a mock SES service for testing
func createMockSESService() (*SESService, *mocks.MockSESClient, *mocks.MockSNSClient, *mocks.MockAuthService) {
	ctrl := gomock.NewController(&testing.T{})
	mockSES := mocks.NewMockSESClient(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := &pkgmocks.MockLogger{}

	return NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	), mockSES, mockSNS, mockAuth
}

// TestListConfigurationSets demonstrates how to test the ListConfigurationSets method
func TestListConfigurationSets(t *testing.T) {
	// Create mock service and clients
	service, mockSESClient, _, _ := createMockSESService()

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	// Setup expectations
	mockOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{
			{Name: aws.String("config-set-1")},
			{Name: aws.String("config-set-2")},
		},
	}

	// Configure the mock to return our test data
	mockSESClient.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockOutput, nil)

	// Call the method being tested
	result, err := service.ListConfigurationSets(context.Background(), testConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Contains(t, result, "config-set-1")
	assert.Contains(t, result, "config-set-2")
}

// TestCreateSNSTopic demonstrates how to test the CreateSNSTopic method
func TestCreateSNSTopic(t *testing.T) {
	// Create mock service and clients
	service, _, mockSNSClient, _ := createMockSESService()

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	topicConfig := domain.SESTopicConfig{
		TopicName:            "test-topic",
		Protocol:             "https",
		NotificationEndpoint: "https://example.com/webhook",
	}

	// Setup expectations
	mockCreateOutput := &sns.CreateTopicOutput{
		TopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:test-topic"),
	}
	mockSubscribeOutput := &sns.SubscribeOutput{
		SubscriptionArn: aws.String("arn:aws:sns:us-east-1:123456789012:test-topic:subscription-id"),
	}

	// Configure the mocks
	mockSNSClient.EXPECT().
		CreateTopicWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.CreateTopicInput, _ ...request.Option) (*sns.CreateTopicOutput, error) {
			assert.Equal(t, topicConfig.TopicName, *input.Name)
			return mockCreateOutput, nil
		})

	mockSNSClient.EXPECT().
		SubscribeWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.SubscribeInput, _ ...request.Option) (*sns.SubscribeOutput, error) {
			assert.Equal(t, topicConfig.Protocol, *input.Protocol)
			assert.Equal(t, topicConfig.NotificationEndpoint, *input.Endpoint)
			assert.Equal(t, *mockCreateOutput.TopicArn, *input.TopicArn)
			return mockSubscribeOutput, nil
		})

	// Call the method being tested
	result, err := service.CreateSNSTopic(context.Background(), testConfig, topicConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, *mockCreateOutput.TopicArn, result)
}

// TestRegisterWebhooks demonstrates how to test the RegisterWebhooks method
func TestRegisterWebhooks(t *testing.T) {
	// Create mock service and clients
	service, mockSESClient, mockSNSClient, _ := createMockSESService()

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"
	baseURL := "https://example.com"
	eventTypes := []domain.EmailEventType{domain.EmailEventDelivered, domain.EmailEventBounce}

	providerConfig := &domain.EmailProvider{
		SES: &domain.AmazonSESSettings{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Region:    "us-east-1",
		},
	}

	// Expected configuration set and topic names
	configSetName := "notifuse-test-integration"
	topicARN := "arn:aws:sns:us-east-1:123456789012:notifuse-ses-test-integration"

	// Setup mock outputs
	mockListConfigOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{},
	}

	mockCreateConfigOutput := &ses.CreateConfigurationSetOutput{}

	mockCreateTopicOutput := &sns.CreateTopicOutput{
		TopicArn: aws.String(topicARN),
	}

	mockSubscribeOutput := &sns.SubscribeOutput{}

	mockListEventDestOutput := &ses.DescribeConfigurationSetOutput{
		EventDestinations: []*ses.EventDestination{},
	}

	mockCreateEventDestOutput := &ses.CreateConfigurationSetEventDestinationOutput{}

	// Configure mocks
	mockSESClient.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockListConfigOutput, nil)

	mockSESClient.EXPECT().
		CreateConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.CreateConfigurationSetInput, _ ...request.Option) (*ses.CreateConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSet.Name)
			return mockCreateConfigOutput, nil
		})

	mockSNSClient.EXPECT().
		CreateTopicWithContext(gomock.Any(), gomock.Any()).
		Return(mockCreateTopicOutput, nil)

	mockSNSClient.EXPECT().
		SubscribeWithContext(gomock.Any(), gomock.Any()).
		Return(mockSubscribeOutput, nil)

	mockSESClient.EXPECT().
		DescribeConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.DescribeConfigurationSetInput, _ ...request.Option) (*ses.DescribeConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			return mockListEventDestOutput, nil
		})

	mockSESClient.EXPECT().
		CreateConfigurationSetEventDestinationWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.CreateConfigurationSetEventDestinationInput, _ ...request.Option) (*ses.CreateConfigurationSetEventDestinationOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			assert.Equal(t, topicARN, *input.EventDestination.SNSDestination.TopicARN)
			return mockCreateEventDestOutput, nil
		})

	// Call the method being tested
	result, err := service.RegisterWebhooks(context.Background(), workspaceID, integrationID, baseURL, eventTypes, providerConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, domain.EmailProviderKindSES, result.EmailProviderKind)
	assert.True(t, result.IsRegistered)
	assert.Len(t, result.Endpoints, 2)
}

// TestNewSESServiceMinimal tests the NewSESService constructor
func TestNewSESServiceMinimal(t *testing.T) {
	// Set up mock auth service
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESService(mockAuth, mockLogger)

	// Verify
	assert.NotNil(t, service)
	assert.NotNil(t, service.sessionFactory)
	assert.NotNil(t, service.sesClientFactory)
	assert.NotNil(t, service.snsClientFactory)
	assert.Equal(t, mockAuth, service.authService)
	assert.Equal(t, mockLogger, service.logger)
}

// TestGetClientsInvalidCredentialsMinimal tests the getClients method with invalid credentials
func TestGetClientsInvalidCredentialsMinimal(t *testing.T) {
	// Set up mock auth service and logger
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service with no factories since they shouldn't be called
	service := &SESService{
		authService: mockAuth,
		logger:      mockLogger,
	}

	// Test with empty credentials
	emptyConfig := domain.AmazonSESSettings{
		AccessKey: "",
		SecretKey: "test-secret",
		Region:    "us-east-1",
	}

	// Call the method
	sesClient, snsClient, err := service.getClients(emptyConfig)

	// Verify results
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidAWSCredentials, err)
	assert.Nil(t, sesClient)
	assert.Nil(t, snsClient)
}

// TestDeleteConfigurationSetMinimal tests the DeleteConfigurationSet method
func TestDeleteConfigurationSetMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	configSetName := "test-config-set"

	// Setup mock outputs
	mockDeleteOutput := &ses.DeleteConfigurationSetOutput{}

	// Configure mock
	mockSES.EXPECT().
		DeleteConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.DeleteConfigurationSetInput, _ ...request.Option) (*ses.DeleteConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			return mockDeleteOutput, nil
		})

	// Call the method being tested
	err := service.DeleteConfigurationSet(context.Background(), testConfig, configSetName)

	// Assert the results
	assert.NoError(t, err)
}

// TestUpdateEventDestinationMinimal tests the UpdateEventDestination method
func TestUpdateEventDestinationMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	destination := domain.SESConfigurationSetEventDestination{
		ConfigurationSetName: "test-config-set",
		Name:                 "test-destination",
		Enabled:              true,
		MatchingEventTypes:   []string{"Delivery", "Bounce"},
		SNSDestination: &domain.SESTopicConfig{
			TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
		},
	}

	// Setup mock outputs
	mockUpdateOutput := &ses.UpdateConfigurationSetEventDestinationOutput{}

	// Configure mock
	mockSES.EXPECT().
		UpdateConfigurationSetEventDestinationWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.UpdateConfigurationSetEventDestinationInput, _ ...request.Option) (*ses.UpdateConfigurationSetEventDestinationOutput, error) {
			assert.Equal(t, destination.ConfigurationSetName, *input.ConfigurationSetName)
			assert.Equal(t, destination.Name, *input.EventDestination.Name)
			assert.Equal(t, destination.Enabled, *input.EventDestination.Enabled)
			assert.Equal(t, destination.SNSDestination.TopicARN, *input.EventDestination.SNSDestination.TopicARN)
			assert.Len(t, input.EventDestination.MatchingEventTypes, 2)
			return mockUpdateOutput, nil
		})

	// Call the method being tested
	err := service.UpdateEventDestination(context.Background(), testConfig, destination)

	// Assert the results
	assert.NoError(t, err)
}

// TestCreateSNSTopicWithNewTopicMinimal tests the CreateSNSTopic method with a new topic
func TestCreateSNSTopicWithNewTopicMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	topicConfig := domain.SESTopicConfig{
		TopicName:            "test-topic-name",
		Protocol:             "https",
		NotificationEndpoint: "https://example.com/webhook",
	}

	expectedTopicARN := "arn:aws:sns:us-east-1:123456789012:test-topic-name"

	// 1. First it should try to check if the topic exists with GetTopicAttributes - should return error because no ARN is provided
	mockSNS.EXPECT().
		GetTopicAttributesWithContext(gomock.Any(), gomock.Any()).
		Times(0)

	// 2. Then it should create the topic
	mockSNS.EXPECT().
		CreateTopicWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.CreateTopicInput, _ ...request.Option) (*sns.CreateTopicOutput, error) {
			assert.Equal(t, topicConfig.TopicName, *input.Name)
			return &sns.CreateTopicOutput{
				TopicArn: aws.String(expectedTopicARN),
			}, nil
		})

	// 3. It also calls SubscribeWithContext
	mockSNS.EXPECT().
		SubscribeWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.SubscribeInput, _ ...request.Option) (*sns.SubscribeOutput, error) {
			assert.Equal(t, expectedTopicARN, *input.TopicArn)
			assert.Equal(t, topicConfig.Protocol, *input.Protocol)
			assert.Equal(t, topicConfig.NotificationEndpoint, *input.Endpoint)
			return &sns.SubscribeOutput{
				SubscriptionArn: aws.String("arn:aws:sns:us-east-1:123456789012:test-topic-name:subscription-id"),
			}, nil
		})

	// Call the method being tested
	result, err := service.CreateSNSTopic(context.Background(), testConfig, topicConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, expectedTopicARN, result)
}

// TestCreateSNSTopicWithExistingARNMinimal tests CreateSNSTopic with an existing topic ARN
func TestCreateSNSTopicWithExistingARNMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	existingTopicARN := "arn:aws:sns:us-east-1:123456789012:existing-topic"
	topicConfig := domain.SESTopicConfig{
		TopicARN: existingTopicARN,
	}

	// Setup mock outputs
	mockTopicAttrsOutput := &sns.GetTopicAttributesOutput{
		Attributes: map[string]*string{
			"TopicArn": aws.String(existingTopicARN),
		},
	}

	// Configure mock to verify topic existence
	mockSNS.EXPECT().
		GetTopicAttributesWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.GetTopicAttributesInput, _ ...request.Option) (*sns.GetTopicAttributesOutput, error) {
			assert.Equal(t, existingTopicARN, *input.TopicArn)
			return mockTopicAttrsOutput, nil
		})

	// Call the method being tested
	result, err := service.CreateSNSTopic(context.Background(), testConfig, topicConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, existingTopicARN, result)
}

// TestDeleteSNSTopicMinimal tests the DeleteSNSTopic method
func TestDeleteSNSTopicMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	topicARN := "arn:aws:sns:us-east-1:123456789012:test-topic"

	// Setup mock outputs
	mockDeleteOutput := &sns.DeleteTopicOutput{}

	// Configure mock
	mockSNS.EXPECT().
		DeleteTopicWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *sns.DeleteTopicInput, _ ...request.Option) (*sns.DeleteTopicOutput, error) {
			assert.Equal(t, topicARN, *input.TopicArn)
			return mockDeleteOutput, nil
		})

	// Call the method being tested
	err := service.DeleteSNSTopic(context.Background(), testConfig, topicARN)

	// Assert the results
	assert.NoError(t, err)
}

// TestCreateConfigurationSetMinimal tests the CreateConfigurationSet method
func TestCreateConfigurationSetMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	configSetName := "test-config-set"

	// Setup mock outputs
	mockCreateOutput := &ses.CreateConfigurationSetOutput{}

	// Configure mock
	mockSES.EXPECT().
		CreateConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.CreateConfigurationSetInput, _ ...request.Option) (*ses.CreateConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSet.Name)
			return mockCreateOutput, nil
		})

	// Call the method being tested
	err := service.CreateConfigurationSet(context.Background(), testConfig, configSetName)

	// Assert the results
	assert.NoError(t, err)
}

// TestListEventDestinationsMinimal tests the ListEventDestinations method
func TestListEventDestinationsMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	configSetName := "test-config-set"
	topicARN := "arn:aws:sns:us-east-1:123456789012:test-topic"

	// Setup mock outputs with event destinations
	mockDescribeOutput := &ses.DescribeConfigurationSetOutput{
		EventDestinations: []*ses.EventDestination{
			{
				Name:    aws.String("test-destination-1"),
				Enabled: aws.Bool(true),
				MatchingEventTypes: []*string{
					aws.String("send"),
					aws.String("delivery"),
				},
				SNSDestination: &ses.SNSDestination{
					TopicARN: aws.String(topicARN),
				},
			},
			{
				Name:    aws.String("test-destination-2"),
				Enabled: aws.Bool(false),
				MatchingEventTypes: []*string{
					aws.String("bounce"),
					aws.String("complaint"),
				},
				SNSDestination: &ses.SNSDestination{
					TopicARN: aws.String(topicARN),
				},
			},
		},
	}

	// Configure mock
	mockSES.EXPECT().
		DescribeConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.DescribeConfigurationSetInput, _ ...request.Option) (*ses.DescribeConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			return mockDescribeOutput, nil
		})

	// Call the method being tested
	destinations, err := service.ListEventDestinations(context.Background(), testConfig, configSetName)

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, destinations, 2)

	// Check first destination
	assert.Equal(t, "test-destination-1", destinations[0].Name)
	assert.Equal(t, configSetName, destinations[0].ConfigurationSetName)
	assert.True(t, destinations[0].Enabled)
	assert.Equal(t, []string{"send", "delivery"}, destinations[0].MatchingEventTypes)
	assert.NotNil(t, destinations[0].SNSDestination)
	assert.Equal(t, topicARN, destinations[0].SNSDestination.TopicARN)

	// Check second destination
	assert.Equal(t, "test-destination-2", destinations[1].Name)
	assert.Equal(t, configSetName, destinations[1].ConfigurationSetName)
	assert.False(t, destinations[1].Enabled)
	assert.Equal(t, []string{"bounce", "complaint"}, destinations[1].MatchingEventTypes)
	assert.NotNil(t, destinations[1].SNSDestination)
	assert.Equal(t, topicARN, destinations[1].SNSDestination.TopicARN)
}

// TestCreateEventDestinationMinimal tests the CreateEventDestination method
func TestCreateEventDestinationMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	destination := domain.SESConfigurationSetEventDestination{
		ConfigurationSetName: "test-config-set",
		Name:                 "test-destination",
		Enabled:              true,
		MatchingEventTypes:   []string{"send", "delivery", "bounce", "complaint"},
		SNSDestination: &domain.SESTopicConfig{
			TopicARN: "arn:aws:sns:us-east-1:123456789012:test-topic",
		},
	}

	// Setup mock outputs
	mockCreateOutput := &ses.CreateConfigurationSetEventDestinationOutput{}

	// Configure mock
	mockSES.EXPECT().
		CreateConfigurationSetEventDestinationWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.CreateConfigurationSetEventDestinationInput, _ ...request.Option) (*ses.CreateConfigurationSetEventDestinationOutput, error) {
			assert.Equal(t, destination.ConfigurationSetName, *input.ConfigurationSetName)
			assert.Equal(t, destination.Name, *input.EventDestination.Name)
			assert.Equal(t, destination.Enabled, *input.EventDestination.Enabled)
			assert.Equal(t, destination.SNSDestination.TopicARN, *input.EventDestination.SNSDestination.TopicARN)
			assert.Len(t, input.EventDestination.MatchingEventTypes, 4)
			return mockCreateOutput, nil
		})

	// Call the method being tested
	err := service.CreateEventDestination(context.Background(), testConfig, destination)

	// Assert the results
	assert.NoError(t, err)
}

// TestCreateEventDestinationErrorMinimal tests the CreateEventDestination method with error
func TestCreateEventDestinationErrorMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}

	// Missing SNS destination
	invalidDestination := domain.SESConfigurationSetEventDestination{
		ConfigurationSetName: "test-config-set",
		Name:                 "test-destination",
		Enabled:              true,
		MatchingEventTypes:   []string{"send", "delivery"},
		SNSDestination:       nil, // Missing SNS destination
	}

	// Call the method being tested
	err := service.CreateEventDestination(context.Background(), testConfig, invalidDestination)

	// Assert the results
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSNSDestination, err)
}

// TestDeleteEventDestinationMinimal tests the DeleteEventDestination method
func TestDeleteEventDestinationMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	testConfig := domain.AmazonSESSettings{
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
		Region:    "us-east-1",
	}
	configSetName := "test-config-set"
	destinationName := "test-destination"

	// Setup mock outputs
	mockDeleteOutput := &ses.DeleteConfigurationSetEventDestinationOutput{}

	// Configure mock
	mockSES.EXPECT().
		DeleteConfigurationSetEventDestinationWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.DeleteConfigurationSetEventDestinationInput, _ ...request.Option) (*ses.DeleteConfigurationSetEventDestinationOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			assert.Equal(t, destinationName, *input.EventDestinationName)
			return mockDeleteOutput, nil
		})

	// Call the method being tested
	err := service.DeleteEventDestination(context.Background(), testConfig, configSetName, destinationName)

	// Assert the results
	assert.NoError(t, err)
}

// TestGetWebhookStatusMinimal tests the GetWebhookStatus method
func TestGetWebhookStatusMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"
	configSetName := "notifuse-test-integration"
	topicARN := "arn:aws:sns:us-east-1:123456789012:notifuse-ses-test-integration"

	providerConfig := &domain.EmailProvider{
		SES: &domain.AmazonSESSettings{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Region:    "us-east-1",
		},
	}

	// Setup mock outputs for ListConfigurationSets and DescribeConfigurationSet
	mockListConfigOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{
			{Name: aws.String(configSetName)},
		},
	}

	mockDescribeOutput := &ses.DescribeConfigurationSetOutput{
		EventDestinations: []*ses.EventDestination{
			{
				Name:    aws.String("notifuse-events"),
				Enabled: aws.Bool(true),
				MatchingEventTypes: []*string{
					aws.String("send"),
					aws.String("delivery"),
					aws.String("bounce"),
					aws.String("complaint"),
				},
				SNSDestination: &ses.SNSDestination{
					TopicARN: aws.String(topicARN),
				},
			},
		},
	}

	// Configure mocks
	mockSES.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockListConfigOutput, nil)

	mockSES.EXPECT().
		DescribeConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, input *ses.DescribeConfigurationSetInput, _ ...request.Option) (*ses.DescribeConfigurationSetOutput, error) {
			assert.Equal(t, configSetName, *input.ConfigurationSetName)
			return mockDescribeOutput, nil
		})

	// Call the method being tested
	status, err := service.GetWebhookStatus(context.Background(), workspaceID, integrationID, providerConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, domain.EmailProviderKindSES, status.EmailProviderKind)
	assert.True(t, status.IsRegistered)
	assert.GreaterOrEqual(t, len(status.Endpoints), 3)

	// Check that endpoints contain expected event types
	eventTypes := make(map[string]bool)
	for _, endpoint := range status.Endpoints {
		eventTypes[string(endpoint.EventType)] = true
	}
	assert.True(t, eventTypes["delivered"])
	assert.True(t, eventTypes["bounce"])
	assert.True(t, eventTypes["complaint"])
}

// TestGetWebhookStatusNotRegisteredMinimal tests the GetWebhookStatus method for unregistered webhooks
func TestGetWebhookStatusNotRegisteredMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"

	providerConfig := &domain.EmailProvider{
		SES: &domain.AmazonSESSettings{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Region:    "us-east-1",
		},
	}

	// Setup mock output for empty configuration sets list
	mockListConfigOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{},
	}

	// Configure mock to return empty config sets list
	mockSES.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockListConfigOutput, nil)

	// Call the method being tested
	status, err := service.GetWebhookStatus(context.Background(), workspaceID, integrationID, providerConfig)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, domain.EmailProviderKindSES, status.EmailProviderKind)
	assert.False(t, status.IsRegistered)
	assert.Empty(t, status.Endpoints)
}

// TestUnregisterWebhooksMinimal tests the UnregisterWebhooks method
func TestUnregisterWebhooksMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"
	configSetName := "notifuse-test-integration"
	topicARN := "arn:aws:sns:us-east-1:123456789012:notifuse-ses-test-integration"
	eventDestinationName := "notifuse-events"

	providerConfig := &domain.EmailProvider{
		SES: &domain.AmazonSESSettings{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Region:    "us-east-1",
		},
	}

	// Setup mock outputs for ListConfigurationSets, DescribeConfigurationSet, and delete operations
	mockListConfigOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{
			{Name: aws.String(configSetName)},
		},
	}

	mockDescribeOutput := &ses.DescribeConfigurationSetOutput{
		EventDestinations: []*ses.EventDestination{
			{
				Name:    aws.String(eventDestinationName),
				Enabled: aws.Bool(true),
				MatchingEventTypes: []*string{
					aws.String("send"),
					aws.String("delivery"),
				},
				SNSDestination: &ses.SNSDestination{
					TopicARN: aws.String(topicARN),
				},
			},
		},
	}

	// Configure mocks with allowances for any order
	// List config sets
	mockSES.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockListConfigOutput, nil).
		AnyTimes()

	// Describe config set to get event destinations
	mockSES.EXPECT().
		DescribeConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		Return(mockDescribeOutput, nil).
		AnyTimes()

	// Delete event destination
	mockSES.EXPECT().
		DeleteConfigurationSetEventDestinationWithContext(gomock.Any(), gomock.Any()).
		Return(&ses.DeleteConfigurationSetEventDestinationOutput{}, nil).
		AnyTimes()

	// Delete configuration set
	mockSES.EXPECT().
		DeleteConfigurationSetWithContext(gomock.Any(), gomock.Any()).
		Return(&ses.DeleteConfigurationSetOutput{}, nil).
		AnyTimes()

	// Delete SNS topic
	mockSNS.EXPECT().
		DeleteTopicWithContext(gomock.Any(), gomock.Any()).
		Return(&sns.DeleteTopicOutput{}, nil).
		AnyTimes()

	// Call the method being tested
	err := service.UnregisterWebhooks(context.Background(), workspaceID, integrationID, providerConfig)

	// Assert the results
	assert.NoError(t, err)
}

// TestUnregisterWebhooksNotRegisteredMinimal tests the UnregisterWebhooks method when webhooks aren't registered
func TestUnregisterWebhooksNotRegisteredMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"

	providerConfig := &domain.EmailProvider{
		SES: &domain.AmazonSESSettings{
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			Region:    "us-east-1",
		},
	}

	// Setup mock output for empty configuration sets list
	mockListConfigOutput := &ses.ListConfigurationSetsOutput{
		ConfigurationSets: []*ses.ConfigurationSet{},
	}

	// Configure mock to return empty config sets list
	mockSES.EXPECT().
		ListConfigurationSetsWithContext(gomock.Any(), gomock.Any()).
		Return(mockListConfigOutput, nil)

	// Call the method being tested
	err := service.UnregisterWebhooks(context.Background(), workspaceID, integrationID, providerConfig)

	// Assert the results - no error when webhooks aren't registered
	assert.NoError(t, err)
}

// TestUnregisterWebhooksInvalidProviderMinimal tests the UnregisterWebhooks method with invalid provider
func TestUnregisterWebhooksInvalidProviderMinimal(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)
	mockSNS := mocks.NewMockSNSClient(ctrl)
	mockSES := mocks.NewMockSESClient(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Create service
	service := NewSESServiceWithClients(
		mockAuth,
		mockLogger,
		func(_ domain.AmazonSESSettings) (*session.Session, error) {
			return &session.Session{}, nil
		},
		func(_ *session.Session) domain.SESWebhookClient {
			return mockSES
		},
		func(_ *session.Session) domain.SNSWebhookClient {
			return mockSNS
		},
	)

	// Setup test data
	workspaceID := "test-workspace"
	integrationID := "test-integration"

	// Invalid provider config (no SES config)
	providerConfig := &domain.EmailProvider{
		Mailgun: &domain.MailgunSettings{},
	}

	// Call the method being tested
	err := service.UnregisterWebhooks(context.Background(), workspaceID, integrationID, providerConfig)

	// Assert the results
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSESConfig, err)
}

// TestSendEmail tests the SendEmail method - using a direct approach with interfaces
func TestSendEmail(t *testing.T) {
	// Set up mock services
	ctrl := gomock.NewController(t)
	mockAuth := mocks.NewMockAuthService(ctrl)
	mockLogger := pkgmocks.NewMockLogger(ctrl)

	// Expectations for logger
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().WithField(gomock.Any(), gomock.Any()).Return(mockLogger).AnyTimes()

	// Test data
	workspaceID := "test-workspace"
	fromAddress := "sender@example.com"
	fromName := "Test Sender"
	to := "recipient@example.com"
	subject := "Test Subject"
	content := "<p>Test Email Content</p>"

	t.Run("missing SES configuration", func(t *testing.T) {
		// Create a regular service
		service := NewSESService(mockAuth, mockLogger)

		// Create provider without SES config
		provider := &domain.EmailProvider{}

		// Call the service method
		err := service.SendEmail(context.Background(), workspaceID, fromAddress, fromName, to, subject, content, provider)

		// Verify error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SES provider is not configured")
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Create a regular service
		service := NewSESService(mockAuth, mockLogger)

		// Create provider with invalid credentials
		provider := &domain.EmailProvider{
			SES: &domain.AmazonSESSettings{
				AccessKey: "", // Empty access key
				SecretKey: "test-secret-key",
				Region:    "us-east-1",
			},
		}

		// Call the service method
		err := service.SendEmail(context.Background(), workspaceID, fromAddress, fromName, to, subject, content, provider)

		// Verify error
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAWSCredentials, err)
	})
}
