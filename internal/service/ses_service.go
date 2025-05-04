package service

import (
	"context"
	"fmt"

	"github.com/Notifuse/notifuse/internal/domain"
	"github.com/Notifuse/notifuse/pkg/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
)

// SESService implements the domain.SESServiceInterface
type SESService struct {
	authService domain.AuthService
	logger      logger.Logger
}

// NewSESService creates a new instance of SESService
func NewSESService(authService domain.AuthService, logger logger.Logger) *SESService {
	return &SESService{
		authService: authService,
		logger:      logger,
	}
}

// createSession creates an AWS session with the given configuration
func createSession(config domain.SESConfig) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
	})
}

// ListConfigurationSets lists all configuration sets
func (s *SESService) ListConfigurationSets(ctx context.Context, config domain.SESConfig) ([]string, error) {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// List configuration sets
	input := &ses.ListConfigurationSetsInput{}
	result, err := svc.ListConfigurationSetsWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list SES configuration sets: %v", err))
		return nil, fmt.Errorf("failed to list SES configuration sets: %w", err)
	}

	// Extract configuration set names
	var configSets []string
	for _, configSet := range result.ConfigurationSets {
		configSets = append(configSets, *configSet.Name)
	}

	return configSets, nil
}

// CreateConfigurationSet creates a new configuration set
func (s *SESService) CreateConfigurationSet(ctx context.Context, config domain.SESConfig, name string) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// Create configuration set
	input := &ses.CreateConfigurationSetInput{
		ConfigurationSet: &ses.ConfigurationSet{
			Name: aws.String(name),
		},
	}

	_, err = svc.CreateConfigurationSetWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create SES configuration set: %v", err))
		return fmt.Errorf("failed to create SES configuration set: %w", err)
	}

	return nil
}

// DeleteConfigurationSet deletes a configuration set
func (s *SESService) DeleteConfigurationSet(ctx context.Context, config domain.SESConfig, name string) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// Delete configuration set
	input := &ses.DeleteConfigurationSetInput{
		ConfigurationSetName: aws.String(name),
	}

	_, err = svc.DeleteConfigurationSetWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete SES configuration set: %v", err))
		return fmt.Errorf("failed to delete SES configuration set: %w", err)
	}

	return nil
}

// CreateSNSTopic creates a new SNS topic for notifications
func (s *SESService) CreateSNSTopic(ctx context.Context, config domain.SESConfig, topicConfig domain.SESTopicConfig) (string, error) {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return "", fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SNS client
	svc := sns.New(sess)

	// If a topic ARN is provided, check if it exists
	if topicConfig.TopicARN != "" {
		// Check if the topic exists
		_, err := svc.GetTopicAttributesWithContext(ctx, &sns.GetTopicAttributesInput{
			TopicArn: aws.String(topicConfig.TopicARN),
		})
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to get SNS topic attributes: %v", err))
			return "", fmt.Errorf("failed to get SNS topic attributes: %w", err)
		}
		return topicConfig.TopicARN, nil
	}

	// Create a new SNS topic if no ARN was provided
	topicName := topicConfig.TopicName
	if topicName == "" {
		topicName = "notifuse-email-webhooks"
	}

	createResult, err := svc.CreateTopicWithContext(ctx, &sns.CreateTopicInput{
		Name: aws.String(topicName),
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create SNS topic: %v", err))
		return "", fmt.Errorf("failed to create SNS topic: %w", err)
	}

	topicARN := *createResult.TopicArn

	// Configure the SNS subscription for the webhook endpoint
	_, err = svc.SubscribeWithContext(ctx, &sns.SubscribeInput{
		Protocol: aws.String(topicConfig.Protocol),
		TopicArn: aws.String(topicARN),
		Endpoint: aws.String(topicConfig.NotificationEndpoint),
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create SNS subscription: %v", err))
		return "", fmt.Errorf("failed to create SNS subscription: %w", err)
	}

	return topicARN, nil
}

// DeleteSNSTopic deletes an SNS topic
func (s *SESService) DeleteSNSTopic(ctx context.Context, config domain.SESConfig, topicARN string) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SNS client
	svc := sns.New(sess)

	// Delete the SNS topic
	_, err = svc.DeleteTopicWithContext(ctx, &sns.DeleteTopicInput{
		TopicArn: aws.String(topicARN),
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete SNS topic: %v", err))
		return fmt.Errorf("failed to delete SNS topic: %w", err)
	}

	return nil
}

// CreateEventDestination creates an event destination in a configuration set
func (s *SESService) CreateEventDestination(ctx context.Context, config domain.SESConfig, destination domain.SESConfigurationSetEventDestination) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// Validate destination
	if destination.SNSDestination == nil || destination.SNSDestination.TopicARN == "" {
		return fmt.Errorf("SNS destination and Topic ARN are required")
	}

	// Convert event types to SES format
	var eventTypes []*string
	for _, eventType := range destination.MatchingEventTypes {
		eventTypes = append(eventTypes, aws.String(eventType))
	}

	// Create event destination
	input := &ses.CreateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(destination.ConfigurationSetName),
		EventDestination: &ses.EventDestination{
			Name:               aws.String(destination.Name),
			Enabled:            aws.Bool(destination.Enabled),
			MatchingEventTypes: aws.StringSlice(destination.MatchingEventTypes),
			SNSDestination: &ses.SNSDestination{
				TopicARN: aws.String(destination.SNSDestination.TopicARN),
			},
		},
	}

	_, err = svc.CreateConfigurationSetEventDestinationWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create SES event destination: %v", err))
		return fmt.Errorf("failed to create SES event destination: %w", err)
	}

	return nil
}

// UpdateEventDestination updates an event destination
func (s *SESService) UpdateEventDestination(ctx context.Context, config domain.SESConfig, destination domain.SESConfigurationSetEventDestination) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// Convert event types to SES format
	var eventTypes []*string
	for _, eventType := range destination.MatchingEventTypes {
		eventTypes = append(eventTypes, aws.String(eventType))
	}

	// Update event destination
	input := &ses.UpdateConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(destination.ConfigurationSetName),
		EventDestination: &ses.EventDestination{
			Name:               aws.String(destination.Name),
			Enabled:            aws.Bool(destination.Enabled),
			MatchingEventTypes: aws.StringSlice(destination.MatchingEventTypes),
			SNSDestination: &ses.SNSDestination{
				TopicARN: aws.String(destination.SNSDestination.TopicARN),
			},
		},
	}

	_, err = svc.UpdateConfigurationSetEventDestinationWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update SES event destination: %v", err))
		return fmt.Errorf("failed to update SES event destination: %w", err)
	}

	return nil
}

// DeleteEventDestination deletes an event destination
func (s *SESService) DeleteEventDestination(ctx context.Context, config domain.SESConfig, configSetName, destinationName string) error {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// Delete event destination
	input := &ses.DeleteConfigurationSetEventDestinationInput{
		ConfigurationSetName: aws.String(configSetName),
		EventDestinationName: aws.String(destinationName),
	}

	_, err = svc.DeleteConfigurationSetEventDestinationWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete SES event destination: %v", err))
		return fmt.Errorf("failed to delete SES event destination: %w", err)
	}

	return nil
}

// ListEventDestinations lists all event destinations for a configuration set
func (s *SESService) ListEventDestinations(ctx context.Context, config domain.SESConfig, configSetName string) ([]domain.SESConfigurationSetEventDestination, error) {
	// Authenticate user
	workspaceID := "system"
	ctx, _, err := s.authService.AuthenticateUserForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Create AWS session
	sess, err := createSession(config)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to create AWS session: %v", err))
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create SES client
	svc := ses.New(sess)

	// List event destinations
	input := &ses.DescribeConfigurationSetInput{
		ConfigurationSetName: aws.String(configSetName),
		ConfigurationSetAttributeNames: []*string{
			aws.String("eventDestinations"),
		},
	}

	result, err := svc.DescribeConfigurationSetWithContext(ctx, input)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to list SES event destinations: %v", err))
		return nil, fmt.Errorf("failed to list SES event destinations: %w", err)
	}

	// Convert AWS response to domain model
	var destinations []domain.SESConfigurationSetEventDestination
	for _, dest := range result.EventDestinations {
		// Skip if not an SNS destination
		if dest.SNSDestination == nil {
			continue
		}

		destination := domain.SESConfigurationSetEventDestination{
			Name:                 *dest.Name,
			ConfigurationSetName: configSetName,
			Enabled:              *dest.Enabled,
			MatchingEventTypes:   aws.StringValueSlice(dest.MatchingEventTypes),
			SNSDestination: &domain.SESTopicConfig{
				TopicARN: *dest.SNSDestination.TopicARN,
			},
		}

		destinations = append(destinations, destination)
	}

	return destinations, nil
}
