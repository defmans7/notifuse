import { useState, useEffect } from 'react'
import {
  Form,
  Input,
  Switch,
  Button,
  InputNumber,
  Alert,
  Select,
  Modal,
  message,
  Space,
  Descriptions,
  Tag,
  Drawer,
  Dropdown,
  Popconfirm,
  Card,
  Spin,
  Tooltip
} from 'antd'
import {
  EmailProvider,
  EmailProviderKind,
  Workspace,
  Integration,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  DeleteIntegrationRequest,
  IntegrationType
} from '../../services/api/types'
import { workspaceService } from '../../services/api/workspace'
import { emailService } from '../../services/api/email'
import { Section } from './Section'
import {
  faCheck,
  faChevronDown,
  faEnvelope,
  faExclamationTriangle,
  faTerminal
} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  getWebhookStatus,
  registerWebhook,
  WebhookRegistrationStatus
} from '../../services/api/webhook_registration'
import config from '../../config'
import { faPaperPlane, faPenToSquare, faTrashCan } from '@fortawesome/free-regular-svg-icons'

// Provider types that only support transactional emails, not marketing emails
const transactionalEmailOnly: EmailProviderKind[] = ['postmark', 'mailgun']

// Component Props
interface IntegrationsProps {
  workspace: Workspace | null
  onSave: (updatedWorkspace: Workspace) => Promise<void>
  loading: boolean
  isOwner: boolean
}

// EmailIntegration component props
interface EmailIntegrationProps {
  integration: {
    id: string
    name: string
    type: IntegrationType
    email_provider: EmailProvider
    created_at: string
    updated_at: string
  }
  isOwner: boolean
  workspace: Workspace
  getIntegrationPurpose: (id: string) => string[]
  isIntegrationInUse: (id: string) => boolean
  renderProviderIcon: (providerKind: EmailProviderKind, height?: number) => React.ReactNode
  renderProviderSpecificDetails: (provider: EmailProvider) => React.ReactNode
  startEditEmailProvider: (integration: Integration) => void
  startTestEmailProvider: (integrationId: string) => void
  setIntegrationAsDefault: (id: string, purpose: 'marketing' | 'transactional') => Promise<void>
  deleteIntegration: (integrationId: string) => Promise<void>
}

// EmailIntegration component
const EmailIntegration = ({
  integration,
  isOwner,
  workspace,
  getIntegrationPurpose,
  isIntegrationInUse,
  renderProviderIcon,
  renderProviderSpecificDetails,
  startEditEmailProvider,
  startTestEmailProvider,
  setIntegrationAsDefault,
  deleteIntegration
}: EmailIntegrationProps) => {
  const provider = integration.email_provider
  const purposes = getIntegrationPurpose(integration.id)
  const [webhookStatus, setWebhookStatus] = useState<WebhookRegistrationStatus | null>(null)
  const [loadingWebhooks, setLoadingWebhooks] = useState(false)
  const [registrationInProgress, setRegistrationInProgress] = useState(false)

  // Fetch webhook status when component mounts
  useEffect(() => {
    if (workspace?.id && integration?.id) {
      fetchWebhookStatus()
    }
  }, [workspace?.id, integration?.id])

  // Function to fetch webhook status
  const fetchWebhookStatus = async () => {
    if (!workspace?.id || !integration?.id) return

    setLoadingWebhooks(true)
    try {
      const response = await getWebhookStatus({
        workspace_id: workspace.id,
        integration_id: integration.id
      })

      setWebhookStatus(response.status)
    } catch (error) {
      console.error('Failed to fetch webhook status:', error)
    } finally {
      setLoadingWebhooks(false)
    }
  }

  // Function to register webhooks
  const handleRegisterWebhooks = async () => {
    if (!workspace?.id || !integration?.id) return

    setRegistrationInProgress(true)
    try {
      await registerWebhook({
        workspace_id: workspace.id,
        integration_id: integration.id,
        base_url: config.API_ENDPOINT
      })

      // Refresh webhook status after registration
      await fetchWebhookStatus()
      message.success('Webhooks registered successfully')
    } catch (error) {
      console.error('Failed to register webhooks:', error)
      message.error('Failed to register webhooks')
    } finally {
      setRegistrationInProgress(false)
    }
  }

  // Render webhook status
  const renderWebhookStatus = () => {
    if (loadingWebhooks) {
      return (
        <Descriptions.Item label="Webhooks">
          <Spin size="small" /> Loading webhook status...
        </Descriptions.Item>
      )
    }

    if (!webhookStatus || !webhookStatus.is_registered) {
      return (
        <Descriptions.Item label="Webhooks">
          <div className="mb-2">
            <Tag color="orange">
              <FontAwesomeIcon icon={faExclamationTriangle} className="text-yellow-500 mr-1" />
              delivered
            </Tag>
            <Tag color="orange">
              <FontAwesomeIcon icon={faExclamationTriangle} className="text-yellow-500 mr-1" />
              bounce
            </Tag>
            <Tag color="orange">
              <FontAwesomeIcon icon={faExclamationTriangle} className="text-yellow-500 mr-1" />
              complaint
            </Tag>
          </div>
          {isOwner && (
            <Button
              size="small"
              className="ml-2"
              type="primary"
              onClick={handleRegisterWebhooks}
              loading={registrationInProgress}
            >
              Register Webhooks
            </Button>
          )}
        </Descriptions.Item>
      )
    }

    return (
      <Descriptions.Item label="Webhooks">
        <div>
          {webhookStatus.endpoints && webhookStatus.endpoints.length > 0 && (
            <div className="mb-2">
              {webhookStatus.endpoints.map((endpoint, index) => (
                <span key={index}>
                  <Tooltip title={endpoint.webhook_id + ' - ' + endpoint.url}>
                    <Tag color={endpoint.active ? 'green' : 'orange'}>
                      {endpoint.active ? (
                        <FontAwesomeIcon icon={faCheck} className="text-green-500 mr-1" />
                      ) : (
                        <FontAwesomeIcon
                          icon={faExclamationTriangle}
                          className="text-yellow-500 mr-1"
                        />
                      )}
                      {endpoint.event_type}
                    </Tag>
                  </Tooltip>
                </span>
              ))}
            </div>
          )}

          <div className="mb-2">
            {isOwner && (
              <Popconfirm
                title="Register webhooks?"
                description="This will register or update webhook endpoints for this email provider."
                onConfirm={handleRegisterWebhooks}
                okText="Yes"
                cancelText="No"
              >
                <Button
                  size="small"
                  className="ml-2"
                  type={webhookStatus.is_registered ? undefined : 'primary'}
                  loading={registrationInProgress}
                >
                  {webhookStatus.is_registered ? 'Re-register' : 'Register Webhooks'}
                </Button>
              </Popconfirm>
            )}
          </div>
          {webhookStatus.error && (
            <Alert message={webhookStatus.error} type="error" showIcon className="mt-2" />
          )}
        </div>
      </Descriptions.Item>
    )
  }

  return (
    <Card
      title={
        <>
          <div className="float-right">
            {isOwner ? (
              <Space>
                <Tooltip title="Edit">
                  <Button
                    type="text"
                    onClick={() => startEditEmailProvider(integration)}
                    size="small"
                  >
                    <FontAwesomeIcon icon={faPenToSquare} />
                  </Button>
                </Tooltip>
                <Popconfirm
                  title="Delete this integration?"
                  description="This action cannot be undone."
                  onConfirm={() => deleteIntegration(integration.id)}
                  okText="Yes"
                  cancelText="No"
                >
                  <Tooltip title="Delete">
                    <Button size="small" type="text">
                      <FontAwesomeIcon icon={faTrashCan} />
                    </Button>
                  </Tooltip>
                </Popconfirm>
                <Button onClick={() => startTestEmailProvider(integration.id)} size="small">
                  Test
                </Button>
              </Space>
            ) : null}
          </div>
          {renderProviderIcon(provider.kind, 24)}
        </>
      }
    >
      <Descriptions bordered size="small" column={1} className="mt-2">
        <Descriptions.Item label="Name">{integration.name}</Descriptions.Item>
        <Descriptions.Item label="Sender">
          {provider.default_sender_name} &lt;{provider.default_sender_email}&gt;
        </Descriptions.Item>
        <Descriptions.Item label="Used for">
          <Space>
            {isIntegrationInUse(integration.id) ? (
              <>
                {purposes.includes('Marketing Emails') && (
                  <Tag color="blue">
                    <FontAwesomeIcon icon={faPaperPlane} className="mr-1" /> Marketing Emails
                  </Tag>
                )}
                {purposes.includes('Transactional Emails') && (
                  <Tag color="purple">
                    <FontAwesomeIcon icon={faTerminal} className="mr-1" /> Transactional Emails
                  </Tag>
                )}
                {purposes.length === 0 && <Tag color="red">Not assigned</Tag>}
              </>
            ) : (
              <Tag color="red">Not assigned</Tag>
            )}
            {isOwner && (
              <>
                {!purposes.includes('Marketing Emails') &&
                  !transactionalEmailOnly.includes(provider.kind) && (
                    <Popconfirm
                      title="Set as marketing email provider?"
                      description="All marketing emails (broadcasts, campaigns) will be sent through this provider from now on."
                      onConfirm={() => setIntegrationAsDefault(integration.id, 'marketing')}
                      okText="Yes"
                      cancelText="No"
                    >
                      <Button
                        size="small"
                        className="mr-2 mt-2"
                        type={
                          !workspace?.settings.marketing_email_provider_id ? 'primary' : undefined
                        }
                      >
                        Use for Marketing
                      </Button>
                    </Popconfirm>
                  )}
                {!purposes.includes('Transactional Emails') && (
                  <Popconfirm
                    title="Set as transactional email provider?"
                    description="All transactional emails (notifications, password resets, etc.) will be sent through this provider from now on."
                    onConfirm={() => setIntegrationAsDefault(integration.id, 'transactional')}
                    okText="Yes"
                    cancelText="No"
                  >
                    <Button
                      size="small"
                      className="mt-2"
                      type={
                        !workspace?.settings.transactional_email_provider_id ? 'primary' : undefined
                      }
                    >
                      Use for Transactional
                    </Button>
                  </Popconfirm>
                )}
              </>
            )}
          </Space>
        </Descriptions.Item>
        {renderProviderSpecificDetails(provider)}
        {renderWebhookStatus()}
      </Descriptions>
    </Card>
  )
}

// Helper functions for handling email integrations
// Include existing helper functions from EmailProviderSettings
interface EmailProviderFormValues {
  kind: EmailProviderKind
  ses?: EmailProvider['ses']
  smtp?: EmailProvider['smtp']
  sparkpost?: EmailProvider['sparkpost']
  postmark?: EmailProvider['postmark']
  mailgun?: EmailProvider['mailgun']
  mailjet?: EmailProvider['mailjet']
  default_sender_email: string
  default_sender_name: string
  type?: IntegrationType
}

const constructProviderFromForm = (formValues: EmailProviderFormValues): EmailProvider => {
  const provider: EmailProvider = {
    kind: formValues.kind,
    default_sender_email: formValues.default_sender_email || '',
    default_sender_name: formValues.default_sender_name || 'Default Sender'
  }

  // Add provider-specific settings
  if (formValues.kind === 'ses' && formValues.ses) {
    provider.ses = formValues.ses
  } else if (formValues.kind === 'smtp' && formValues.smtp) {
    provider.smtp = formValues.smtp
  } else if (formValues.kind === 'sparkpost' && formValues.sparkpost) {
    provider.sparkpost = formValues.sparkpost
  } else if (formValues.kind === 'postmark' && formValues.postmark) {
    provider.postmark = formValues.postmark
  } else if (formValues.kind === 'mailgun' && formValues.mailgun) {
    provider.mailgun = formValues.mailgun
  } else if (formValues.kind === 'mailjet' && formValues.mailjet) {
    provider.mailjet = formValues.mailjet
  }

  return provider
}

// Main Integrations component
export function Integrations({ workspace, onSave, loading, isOwner }: IntegrationsProps) {
  // State for providers
  const [emailProviderForm] = Form.useForm()
  const [selectedProviderType, setSelectedProviderType] = useState<EmailProviderKind | null>(null)
  const [editingIntegrationId, setEditingIntegrationId] = useState<string | null>(null)

  // Drawer state
  const [providerDrawerVisible, setProviderDrawerVisible] = useState(false)

  // Test email modal state
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmailAddress, setTestEmailAddress] = useState('')
  const [testingIntegrationId, setTestingIntegrationId] = useState<string | null>(null)
  const [testingProvider, setTestingProvider] = useState<EmailProvider | null>(null)
  const [testingEmailLoading, setTestingEmailLoading] = useState(false)

  if (!workspace) {
    return null
  }

  // Get integration by id
  const getIntegrationById = (id: string): Integration | undefined => {
    return workspace.integrations?.find((i) => i.id === id)
  }

  // Is the integration being used
  const isIntegrationInUse = (id: string): boolean => {
    return (
      workspace.settings.marketing_email_provider_id === id ||
      workspace.settings.transactional_email_provider_id === id
    )
  }

  // Get purpose of integration
  const getIntegrationPurpose = (id: string): string[] => {
    const purposes: string[] = []

    if (workspace.settings.marketing_email_provider_id === id) {
      purposes.push('Marketing Emails')
    }

    if (workspace.settings.transactional_email_provider_id === id) {
      purposes.push('Transactional Emails')
    }

    return purposes
  }

  // Set integration as default for a purpose
  const setIntegrationAsDefault = async (id: string, purpose: 'marketing' | 'transactional') => {
    try {
      const updateData = {
        ...workspace,
        settings: {
          ...workspace.settings,
          ...(purpose === 'marketing'
            ? { marketing_email_provider_id: id }
            : { transactional_email_provider_id: id })
        }
      }

      await workspaceService.update(updateData)

      // Refresh workspace data
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      message.success(`Set as default ${purpose} email provider`)
    } catch (error) {
      console.error('Error setting default provider', error)
      message.error('Failed to set default provider')
    }
  }

  // Start editing an existing email provider
  const startEditEmailProvider = (integration: Integration) => {
    if (integration.type !== 'email') return

    setEditingIntegrationId(integration.id)
    setSelectedProviderType(integration.email_provider.kind)
    emailProviderForm.setFieldsValue({
      name: integration.name,
      ...integration.email_provider
    })
    setProviderDrawerVisible(true)
  }

  // Start testing an email provider
  const startTestEmailProvider = (integrationId: string) => {
    const integration = getIntegrationById(integrationId)
    if (!integration || integration.type !== 'email') {
      message.error('Integration not found or not an email provider')
      return
    }

    setTestingIntegrationId(integrationId)
    setTestingProvider(integration.email_provider)
    setTestEmailAddress('')
    setTestModalVisible(true)
  }

  // Cancel adding/editing email provider
  const cancelEmailProviderOperation = () => {
    closeProviderDrawer()
  }

  // Handle provider selection and open drawer
  const handleSelectProviderType = (provider: EmailProviderKind) => {
    setSelectedProviderType(provider)
    emailProviderForm.setFieldsValue({
      kind: provider,
      type: 'email',
      name: provider.charAt(0).toUpperCase() + provider.slice(1)
    })
    setProviderDrawerVisible(true)
  }

  // Close provider drawer
  const closeProviderDrawer = () => {
    setProviderDrawerVisible(false)
    setSelectedProviderType(null)
    emailProviderForm.resetFields()
  }

  // Save new or edited integration
  const saveEmailProvider = async (values: EmailProviderFormValues & { name?: string }) => {
    if (!workspace) return

    try {
      const provider = constructProviderFromForm(values)
      const name = values.name || provider.kind
      const type: IntegrationType = 'email'

      // If editing an existing integration
      if (editingIntegrationId) {
        const integration = getIntegrationById(editingIntegrationId)
        if (!integration) {
          throw new Error('Integration not found')
        }

        const updateRequest: UpdateIntegrationRequest = {
          workspace_id: workspace.id,
          integration_id: editingIntegrationId,
          name: name,
          provider
        }

        await workspaceService.updateIntegration(updateRequest)
        message.success('Integration updated successfully')
      }
      // Creating a new integration
      else {
        const createRequest: CreateIntegrationRequest = {
          workspace_id: workspace.id,
          name,
          type,
          provider
        }

        await workspaceService.createIntegration(createRequest)
        message.success('Integration created successfully')
      }

      // Refresh workspace data
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      // Reset state
      cancelEmailProviderOperation()
    } catch (error) {
      console.error('Error saving integration', error)
      message.error('Failed to save integration')
    }
  }

  // Delete an integration
  const deleteIntegration = async (integrationId: string) => {
    if (!workspace) return

    try {
      const deleteRequest: DeleteIntegrationRequest = {
        workspace_id: workspace.id,
        integration_id: integrationId
      }

      await workspaceService.deleteIntegration(deleteRequest)

      // Refresh workspace data
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      message.success('Integration deleted successfully')
    } catch (error) {
      console.error('Error deleting integration', error)
      message.error('Failed to delete integration')
    }
  }

  // Handler for testing the email provider
  const handleTestProvider = async () => {
    if (!workspace || !testingProvider || !testEmailAddress) return

    try {
      setTestingEmailLoading(true)

      let providerToTest: EmailProvider

      // If testing an existing integration
      if (testingIntegrationId) {
        const integration = getIntegrationById(testingIntegrationId)
        if (!integration || integration.type !== 'email') {
          message.error('Integration not found or not an email provider')
          return
        }
        providerToTest = integration.email_provider
      } else {
        // Testing a provider that hasn't been saved yet
        providerToTest = testingProvider
      }

      const response = await emailService.testProvider(
        workspace.id,
        providerToTest,
        testEmailAddress
      )

      if (response.success) {
        message.success('Test email sent successfully')
        setTestModalVisible(false)
      } else {
        message.error(`Failed to send test email: ${response.error}`)
      }
    } catch (error) {
      console.error('Error testing email provider', error)
      message.error('Failed to test email provider')
    } finally {
      setTestingEmailLoading(false)
    }
  }

  // Render the list of available integrations
  const renderAvailableIntegrations = () => {
    // Available provider configurations
    const providers = [
      {
        type: 'email' as IntegrationType,
        kind: 'smtp',
        name: 'SMTP',
        icon: <FontAwesomeIcon icon={faEnvelope} className="w-16" />
      },
      {
        type: 'email' as IntegrationType,
        kind: 'ses',
        name: 'Amazon SES',
        icon: <img src="/amazonses.png" alt="Amazon SES" className="h-8 w-16 object-contain" />
      },
      {
        type: 'email' as IntegrationType,
        kind: 'sparkpost',
        name: 'SparkPost',
        icon: <img src="/sparkpost.png" alt="SparkPost" className="h-8 w-16 object-contain" />
      },
      {
        type: 'email' as IntegrationType,
        kind: 'postmark',
        name: 'Postmark',
        icon: <img src="/postmark.png" alt="Postmark" className="h-8 w-16 object-contain" />
      },
      {
        type: 'email' as IntegrationType,
        kind: 'mailgun',
        name: 'Mailgun',
        icon: <img src="/mailgun.png" alt="Mailgun" className="h-8 w-16 object-contain" />
      },
      {
        type: 'email' as IntegrationType,
        kind: 'mailjet',
        name: 'Mailjet',
        icon: <img src="/mailjet.png" alt="Mailjet" className="h-8 w-16 object-contain" />
      }
      // Future integration types can be added here
    ]

    return (
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {providers.map((provider) => (
          <Card
            key={`${provider.type}-${provider.kind}`}
            hoverable
            onClick={() => handleSelectProviderType(provider.kind as EmailProviderKind)}
            className="flex items-center cursor-pointer"
            extra={
              <Button
                type="primary"
                ghost
                size="small"
                onClick={(e) => {
                  e.stopPropagation()
                  handleSelectProviderType(provider.kind as EmailProviderKind)
                }}
              >
                Configure
              </Button>
            }
          >
            <Card.Meta avatar={provider.icon} title={provider.name} />
          </Card>
        ))}
      </div>
    )
  }

  // Render the list of integrations
  const renderWorkspaceIntegrations = () => {
    if (!workspace?.integrations) {
      return null // We'll handle this case differently in the main render
    }

    return (
      <>
        {workspace?.integrations.map((integration) => {
          if (integration.type === 'email') {
            return (
              <div key={integration.id} className="mb-4">
                <EmailIntegration
                  key={integration.id}
                  integration={integration}
                  isOwner={isOwner}
                  workspace={workspace}
                  getIntegrationPurpose={getIntegrationPurpose}
                  isIntegrationInUse={isIntegrationInUse}
                  renderProviderIcon={renderProviderIcon}
                  renderProviderSpecificDetails={renderProviderSpecificDetails}
                  startEditEmailProvider={startEditEmailProvider}
                  startTestEmailProvider={startTestEmailProvider}
                  setIntegrationAsDefault={setIntegrationAsDefault}
                  deleteIntegration={deleteIntegration}
                />
              </div>
            )
          }

          // Handle other types of integrations here in the future
          return (
            <Card key={integration.id} className="mb-4">
              <Card.Meta title={integration.name} description={`Type: ${integration.type}`} />
            </Card>
          )
        })}
      </>
    )
  }

  // Render provider-specific form fields
  const renderEmailProviderForm = (providerType: EmailProviderKind) => {
    return (
      <>
        <Form.Item name="name" label="Integration Name" rules={[{ required: true }]}>
          <Input placeholder="Enter a name for this integration" disabled={!isOwner} />
        </Form.Item>
        <Form.Item name="default_sender_email" label="Sender Email" rules={[{ required: true }]}>
          <Input placeholder="noreply@yourdomain.com" disabled={!isOwner} />
        </Form.Item>
        <Form.Item name="default_sender_name" label="Sender Name" rules={[{ required: true }]}>
          <Input placeholder="Your Company Name" disabled={!isOwner} />
        </Form.Item>

        {providerType === 'ses' && (
          <>
            <Form.Item name={['ses', 'region']} label="AWS Region" rules={[{ required: true }]}>
              <Input placeholder="us-east-1" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['ses', 'access_key']}
              label="AWS Access Key"
              rules={[{ required: true }]}
            >
              <Input placeholder="Access Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['ses', 'secret_key']} label="AWS Secret Key">
              <Input.Password placeholder="Secret Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['ses', 'sandbox_mode']} valuePropName="checked" label="Sandbox Mode">
              <Switch disabled={!isOwner} />
            </Form.Item>
          </>
        )}

        {providerType === 'smtp' && (
          <>
            <Form.Item name={['smtp', 'host']} label="SMTP Host" rules={[{ required: true }]}>
              <Input placeholder="smtp.yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['smtp', 'port']} label="SMTP Port" rules={[{ required: true }]}>
              <InputNumber min={1} max={65535} placeholder="587" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['smtp', 'username']}
              label="SMTP Username"
              rules={[{ required: true }]}
            >
              <Input placeholder="Username" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['smtp', 'password']} label="SMTP Password">
              <Input.Password placeholder="Password" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['smtp', 'use_tls']} valuePropName="checked" label="Use TLS">
              <Switch defaultChecked disabled={!isOwner} />
            </Form.Item>
          </>
        )}

        {providerType === 'sparkpost' && (
          <>
            <Form.Item
              name={['sparkpost', 'endpoint']}
              label="API Endpoint"
              rules={[{ required: true }]}
            >
              <Select
                placeholder="Select SparkPost endpoint"
                disabled={!isOwner}
                options={[
                  { label: 'SparkPost US', value: 'https://api.sparkpost.com' },
                  { label: 'SparkPost EU', value: 'https://api.eu.sparkpost.com' }
                ]}
              />
            </Form.Item>
            <Form.Item name={['sparkpost', 'api_key']} label="SparkPost API Key">
              <Input.Password placeholder="API Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['sparkpost', 'sandbox_mode']}
              valuePropName="checked"
              label="Sandbox Mode"
            >
              <Switch disabled={!isOwner} />
            </Form.Item>
          </>
        )}

        {providerType === 'postmark' && (
          <Form.Item
            name={['postmark', 'server_token']}
            label="Server Token"
            rules={[{ required: true }]}
          >
            <Input.Password placeholder="Server Token" disabled={!isOwner} />
          </Form.Item>
        )}

        {providerType === 'mailgun' && (
          <>
            <Form.Item name={['mailgun', 'domain']} label="Domain" rules={[{ required: true }]}>
              <Input placeholder="mail.yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['mailgun', 'api_key']} label="API Key" rules={[{ required: true }]}>
              <Input.Password placeholder="API Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['mailgun', 'region']} label="Region" initialValue="US">
              <Select
                placeholder="Select Mailgun Region"
                disabled={!isOwner}
                options={[
                  { label: 'US', value: 'US' },
                  { label: 'EU', value: 'EU' }
                ]}
              />
            </Form.Item>
          </>
        )}

        {providerType === 'mailjet' && (
          <>
            <Form.Item name={['mailjet', 'api_key']} label="API Key" rules={[{ required: true }]}>
              <Input.Password placeholder="API Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['mailjet', 'secret_key']}
              label="Secret Key"
              rules={[{ required: true }]}
            >
              <Input.Password placeholder="Secret Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['mailjet', 'sandbox_mode']}
              valuePropName="checked"
              label="Sandbox Mode"
            >
              <Switch disabled={!isOwner} />
            </Form.Item>
          </>
        )}
      </>
    )
  }

  // Render provider icon
  const renderProviderIcon = (providerKind: EmailProviderKind, height?: number) => {
    if (providerKind === 'smtp') {
      return <FontAwesomeIcon icon={faEnvelope} className="h-4 w-8 object-contain" />
    } else if (providerKind === 'ses') {
      return <img src="/amazonses.png" alt="Amazon SES" style={{ height: height || 16 }} />
    } else if (providerKind === 'sparkpost') {
      return <img src="/sparkpost.png" alt="SparkPost" style={{ height: height || 16 }} />
    } else if (providerKind === 'postmark') {
      return <img src="/postmark.png" alt="Postmark" style={{ height: height || 16 }} />
    } else if (providerKind === 'mailgun') {
      return <img src="/mailgun.png" alt="Mailgun" style={{ height: height || 16 }} />
    } else if (providerKind === 'mailjet') {
      return <img src="/mailjet.png" alt="Mailjet" style={{ height: height || 16 }} />
    }
    return null
  }

  // Render provider-specific details in description list
  const renderProviderSpecificDetails = (provider: EmailProvider) => {
    const items = []

    if (provider.kind === 'smtp' && provider.smtp) {
      items.push(
        <Descriptions.Item key="host" label="SMTP Host">
          {provider.smtp.host}:{provider.smtp.port}
        </Descriptions.Item>,
        <Descriptions.Item key="username" label="SMTP User">
          {provider.smtp.username}
        </Descriptions.Item>,
        <Descriptions.Item key="tls" label="TLS Enabled">
          {provider.smtp.use_tls ? 'Yes' : 'No'}
        </Descriptions.Item>
      )
    } else if (provider.kind === 'ses' && provider.ses) {
      items.push(
        <Descriptions.Item key="region" label="AWS Region">
          {provider.ses.region}
        </Descriptions.Item>,
        <Descriptions.Item key="sandbox" label="Sandbox Mode">
          {provider.ses.sandbox_mode ? 'Enabled' : 'Disabled'}
        </Descriptions.Item>
      )
    } else if (provider.kind === 'sparkpost' && provider.sparkpost) {
      items.push(
        <Descriptions.Item key="endpoint" label="API Endpoint">
          {provider.sparkpost.endpoint}
        </Descriptions.Item>,
        <Descriptions.Item key="sandbox" label="Sandbox Mode">
          {provider.sparkpost.sandbox_mode ? 'Enabled' : 'Disabled'}
        </Descriptions.Item>
      )
    } else if (provider.kind === 'mailgun' && provider.mailgun) {
      items.push(
        <Descriptions.Item key="domain" label="Domain">
          {provider.mailgun.domain}
        </Descriptions.Item>,
        <Descriptions.Item key="region" label="Region">
          {provider.mailgun.region || 'US'}
        </Descriptions.Item>
      )
    } else if (provider.kind === 'mailjet' && provider.mailjet) {
      items.push(
        <Descriptions.Item key="sandbox" label="Sandbox Mode">
          {provider.mailjet.sandbox_mode ? 'Enabled' : 'Disabled'}
        </Descriptions.Item>
      )
    }

    return items
  }

  // Render the drawer for configuring email providers
  const renderProviderDrawer = () => {
    // Test provider from the drawer
    const handleTestFromDrawer = () => {
      // Validate form fields before proceeding
      emailProviderForm
        .validateFields()
        .then((values) => {
          // Create a temporary provider object from form values
          const tempProvider = constructProviderFromForm(values)

          // Open test modal with the temporary provider
          setTestEmailAddress('')
          setTestingIntegrationId(null) // No integration ID as this is a new provider
          setTestingProvider(tempProvider)
          setTestModalVisible(true)
        })
        .catch((error) => {
          // Form validation failed
          console.error('Validation failed:', error)
          message.error('Please fill in all required fields before testing')
        })
    }

    return (
      <Drawer
        title={
          editingIntegrationId
            ? `Edit ${selectedProviderType?.toUpperCase() || ''} Integration`
            : `Add New ${selectedProviderType?.toUpperCase() || ''} Integration`
        }
        width={600}
        open={providerDrawerVisible}
        onClose={closeProviderDrawer}
        footer={
          <div style={{ textAlign: 'right' }}>
            <Space>
              <Button onClick={closeProviderDrawer}>Cancel</Button>
              <Button onClick={handleTestFromDrawer}>Test Integration</Button>
              <Button type="primary" onClick={() => emailProviderForm.submit()} loading={loading}>
                Save
              </Button>
            </Space>
          </div>
        }
      >
        {selectedProviderType && (
          <Form
            form={emailProviderForm}
            layout="vertical"
            onFinish={saveEmailProvider}
            initialValues={{ kind: selectedProviderType }}
          >
            <Form.Item name="kind" hidden>
              <Input />
            </Form.Item>

            <Alert
              message="Configure Email Provider"
              description="Give your integration a descriptive name and configure the email provider settings. This integration will allow Notifuse to send emails through your email provider."
              type="info"
              showIcon
              style={{ marginBottom: 24 }}
            />

            {renderEmailProviderForm(selectedProviderType)}
          </Form>
        )}
      </Drawer>
    )
  }

  // Add integration dropdown menu items
  const integrationMenuItems = [
    {
      key: 'smtp',
      label: 'SMTP',
      icon: <FontAwesomeIcon icon={faEnvelope} className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('smtp')
    },
    {
      key: 'ses',
      label: 'Amazon SES',
      icon: <img src="/amazonses.png" alt="Amazon SES" className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('ses')
    },
    {
      key: 'sparkpost',
      label: 'SparkPost',
      icon: <img src="/sparkpost.png" alt="SparkPost" className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('sparkpost')
    },
    {
      key: 'postmark',
      label: 'Postmark',
      icon: <img src="/postmark.png" alt="Postmark" className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('postmark')
    },
    {
      key: 'mailgun',
      label: 'Mailgun',
      icon: <img src="/mailgun.png" alt="Mailgun" className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('mailgun')
    },
    {
      key: 'mailjet',
      label: 'Mailjet',
      icon: <img src="/mailjet.png" alt="Mailjet" className="h-6 w-12 object-contain mr-1" />,
      onClick: () => handleSelectProviderType('mailjet')
    }
    // Future integration types can be added here
  ]

  return (
    <Section
      title="Integrations"
      description="Connect and manage external services"
      extra={
        isOwner && (workspace?.integrations?.length ?? 0) > 0 ? (
          <Dropdown menu={{ items: integrationMenuItems }} trigger={['click']}>
            <Button type="primary" size="small" ghost>
              Add Integration <FontAwesomeIcon icon={faChevronDown} />
            </Button>
          </Dropdown>
        ) : null
      }
    >
      {/* Check and display alert for missing email provider configuration */}
      {workspace && (
        <>
          {(!workspace.settings.transactional_email_provider_id ||
            !workspace.settings.marketing_email_provider_id) && (
            <Alert
              message="Email Provider Configuration Needed"
              description={
                <div>
                  {!workspace.settings.transactional_email_provider_id && (
                    <p>
                      Consider connecting a transactional email provider to be able to use
                      transactional emails for account notifications, password resets, and other
                      important system messages.
                    </p>
                  )}
                  {!workspace.settings.marketing_email_provider_id && (
                    <p>
                      Consider connecting a marketing email provider to send newsletters,
                      promotional campaigns, and announcements to engage with your audience.
                    </p>
                  )}
                </div>
              }
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
          )}
        </>
      )}

      {(workspace?.integrations?.length ?? 0) === 0
        ? renderAvailableIntegrations()
        : renderWorkspaceIntegrations()}

      {/* Provider Configuration Drawer */}
      {renderProviderDrawer()}

      {/* Test email modal */}
      <Modal
        title="Test Email Provider"
        open={testModalVisible}
        onCancel={() => setTestModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setTestModalVisible(false)}>
            Cancel
          </Button>,
          <Button
            key="submit"
            type="primary"
            loading={testingEmailLoading}
            onClick={handleTestProvider}
            disabled={!testEmailAddress}
          >
            Send Test Email
          </Button>
        ]}
      >
        <p>Enter an email address to receive a test email:</p>
        <Input
          placeholder="recipient@example.com"
          value={testEmailAddress}
          onChange={(e) => setTestEmailAddress(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <Alert
          message="This will send a real test email to the address provided."
          type="info"
          showIcon
        />
      </Modal>
    </Section>
  )
}
