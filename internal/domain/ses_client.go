package domain

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sns"
)

// SESWebhookClient defines the interface for interacting with AWS SES service
type SESWebhookClient interface {
	CreateConfigurationSetWithContext(ctx aws.Context, input *ses.CreateConfigurationSetInput, opts ...request.Option) (*ses.CreateConfigurationSetOutput, error)
	DeleteConfigurationSetWithContext(ctx aws.Context, input *ses.DeleteConfigurationSetInput, opts ...request.Option) (*ses.DeleteConfigurationSetOutput, error)
	ListConfigurationSetsWithContext(ctx aws.Context, input *ses.ListConfigurationSetsInput, opts ...request.Option) (*ses.ListConfigurationSetsOutput, error)
	DescribeConfigurationSetWithContext(ctx aws.Context, input *ses.DescribeConfigurationSetInput, opts ...request.Option) (*ses.DescribeConfigurationSetOutput, error)
	CreateConfigurationSetEventDestinationWithContext(ctx aws.Context, input *ses.CreateConfigurationSetEventDestinationInput, opts ...request.Option) (*ses.CreateConfigurationSetEventDestinationOutput, error)
	UpdateConfigurationSetEventDestinationWithContext(ctx aws.Context, input *ses.UpdateConfigurationSetEventDestinationInput, opts ...request.Option) (*ses.UpdateConfigurationSetEventDestinationOutput, error)
	DeleteConfigurationSetEventDestinationWithContext(ctx aws.Context, input *ses.DeleteConfigurationSetEventDestinationInput, opts ...request.Option) (*ses.DeleteConfigurationSetEventDestinationOutput, error)
}

// SNSWebhookClient defines the interface for interacting with AWS SNS service
type SNSWebhookClient interface {
	CreateTopicWithContext(ctx aws.Context, input *sns.CreateTopicInput, opts ...request.Option) (*sns.CreateTopicOutput, error)
	DeleteTopicWithContext(ctx aws.Context, input *sns.DeleteTopicInput, opts ...request.Option) (*sns.DeleteTopicOutput, error)
	SubscribeWithContext(ctx aws.Context, input *sns.SubscribeInput, opts ...request.Option) (*sns.SubscribeOutput, error)
	GetTopicAttributesWithContext(ctx aws.Context, input *sns.GetTopicAttributesInput, opts ...request.Option) (*sns.GetTopicAttributesOutput, error)
}
