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
