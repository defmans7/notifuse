import { useState, useEffect } from 'react'
import {
  Card,
  Form,
  Input,
  Switch,
  Button,
  Row,
  Col,
  InputNumber,
  Alert,
  Select,
  Modal,
  message,
  Space,
  Descriptions,
  Tabs,
  Typography,
  Empty
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
import { MailOutlined, DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons'
import { workspaceService } from '../../services/api/workspace'
import { Section } from './Section'

const { Title, Text } = Typography
const { TabPane } = Tabs

// Component Props
interface IntegrationsProps {
  workspace: Workspace | null
  onSave: (updatedWorkspace: Workspace) => Promise<void>
  loading: boolean
  isOwner: boolean
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
  // State for integrations management
  const [activeTab, setActiveTab] = useState('email')
  const [emailIntegrations, setEmailIntegrations] = useState<Integration[]>([])

  // State for adding/editing email providers
  const [addingEmailProvider, setAddingEmailProvider] = useState(false)
  const [editingIntegrationId, setEditingIntegrationId] = useState<string | null>(null)
  const [emailProviderForm] = Form.useForm()
  const [selectedProviderType, setSelectedProviderType] = useState<EmailProviderKind | null>(null)

  // Test email modal state
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmailAddress, setTestEmailAddress] = useState('')
  const [testingIntegrationId, setTestingIntegrationId] = useState<string | null>(null)
  const [testingEmailLoading, setTestingEmailLoading] = useState(false)

  // Load integrations when workspace changes
  useEffect(() => {
    if (workspace && workspace.integrations) {
      const emailProviders = workspace.integrations.filter((i) => i.type === 'email')
      setEmailIntegrations(emailProviders)
    }
  }, [workspace])

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
  const getIntegrationPurpose = (id: string): string => {
    if (workspace.settings.marketing_email_provider_id === id) {
      return 'Marketing Emails'
    }
    if (workspace.settings.transactional_email_provider_id === id) {
      return 'Transactional Emails'
    }
    return 'Not assigned'
  }

  // Set integration as default for a purpose
  const setIntegrationAsDefault = async (id: string, purpose: 'marketing' | 'transactional') => {
    try {
      const updateData = {
        id: workspace.id,
        settings: {
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

  // Start adding a new email provider
  const startAddEmailProvider = () => {
    setSelectedProviderType(null)
    setAddingEmailProvider(true)
    setEditingIntegrationId(null)
    emailProviderForm.resetFields()
  }

  // Start editing an existing email provider
  const startEditEmailProvider = (integration: Integration) => {
    if (integration.type !== 'email') return

    setEditingIntegrationId(integration.id)
    setSelectedProviderType(integration.email_provider.kind)
    emailProviderForm.setFieldsValue(integration.email_provider)
  }

  // Start testing an email provider
  const startTestEmailProvider = (integrationId: string) => {
    const integration = getIntegrationById(integrationId)
    if (!integration || integration.type !== 'email') {
      message.error('Integration not found or not an email provider')
      return
    }

    setTestingIntegrationId(integrationId)
    setTestEmailAddress('')
    setTestModalVisible(true)
  }

  // Cancel adding/editing email provider
  const cancelEmailProviderOperation = () => {
    setAddingEmailProvider(false)
    setEditingIntegrationId(null)
    setSelectedProviderType(null)
    emailProviderForm.resetFields()
  }

  // Select provider type when adding new provider
  const handleSelectProviderType = (provider: EmailProviderKind) => {
    setSelectedProviderType(provider)
    emailProviderForm.setFieldsValue({ kind: provider })
  }

  // Save new or edited email provider
  const saveEmailProvider = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    try {
      const provider = constructProviderFromForm(values)

      // If editing an existing integration
      if (editingIntegrationId) {
        const integration = getIntegrationById(editingIntegrationId)
        if (!integration) {
          throw new Error('Integration not found')
        }

        const updateRequest: UpdateIntegrationRequest = {
          workspace_id: workspace.id,
          integration_id: editingIntegrationId,
          name: integration.name,
          provider
        }

        await workspaceService.updateIntegration(updateRequest)
        message.success('Email provider updated successfully')
      }
      // Creating a new integration
      else {
        const name = `Email Provider (${provider.kind})`
        const createRequest: CreateIntegrationRequest = {
          workspace_id: workspace.id,
          name,
          type: 'email',
          provider
        }

        await workspaceService.createIntegration(createRequest)
        message.success('Email provider created successfully')
      }

      // Refresh workspace data
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      // Reset state
      cancelEmailProviderOperation()
    } catch (error) {
      console.error('Error saving email provider', error)
      message.error('Failed to save email provider')
    }
  }

  // Delete an integration
  const deleteIntegration = async (integrationId: string) => {
    if (!workspace) return

    // Check if integration is in use
    if (isIntegrationInUse(integrationId)) {
      message.error('Cannot delete an integration that is currently in use')
      return
    }

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
    if (!workspace || !testingIntegrationId || !testEmailAddress) return

    try {
      setTestingEmailLoading(true)
      const integration = getIntegrationById(testingIntegrationId)

      if (!integration || integration.type !== 'email') {
        message.error('Integration not found or not an email provider')
        return
      }

      const response = await workspaceService.testEmailProvider({
        provider: integration.email_provider,
        to: testEmailAddress,
        workspace_id: workspace.id
      })

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

  // Render the list of email provider integrations
  const renderEmailIntegrations = () => {
    if (emailIntegrations.length === 0) {
      return (
        <Empty
          description="No email integrations configured"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
        />
      )
    }

    return emailIntegrations.map((integration) => {
      const provider = integration.email_provider
      const isEditing = editingIntegrationId === integration.id
      const purpose = getIntegrationPurpose(integration.id)

      return (
        <Card
          key={integration.id}
          title={integration.name}
          className="mb-4"
          extra={
            isOwner && !isEditing ? (
              <Space>
                <Button
                  icon={<EditOutlined />}
                  onClick={() => startEditEmailProvider(integration)}
                  size="small"
                >
                  Edit
                </Button>
                <Button
                  icon={<DeleteOutlined />}
                  danger
                  onClick={() => deleteIntegration(integration.id)}
                  disabled={isIntegrationInUse(integration.id)}
                  size="small"
                >
                  Delete
                </Button>
              </Space>
            ) : null
          }
        >
          {isEditing ? (
            <Form
              form={emailProviderForm}
              layout="vertical"
              onFinish={saveEmailProvider}
              initialValues={provider || undefined}
            >
              <Form.Item name="kind" hidden>
                <Input />
              </Form.Item>

              {renderEmailProviderForm(provider.kind)}

              <div className="mt-4 flex justify-end">
                <Space>
                  <Button onClick={cancelEmailProviderOperation}>Cancel</Button>
                  <Button type="primary" htmlType="submit" loading={loading}>
                    Save
                  </Button>
                </Space>
              </div>
            </Form>
          ) : (
            <>
              <Descriptions bordered size="small" column={1}>
                <Descriptions.Item label="Type">
                  <Space>
                    {renderProviderIcon(provider.kind)}
                    {provider.kind.toUpperCase()}
                  </Space>
                </Descriptions.Item>
                <Descriptions.Item label="Sender">
                  {provider.default_sender_name} &lt;{provider.default_sender_email}&gt;
                </Descriptions.Item>
                <Descriptions.Item label="Used for">{purpose}</Descriptions.Item>
                {renderProviderSpecificDetails(provider)}
              </Descriptions>

              <div className="mt-4 flex justify-end">
                <Space>
                  <Button onClick={() => startTestEmailProvider(integration.id)}>Test</Button>

                  {purpose !== 'Marketing Emails' && (
                    <Button
                      onClick={() => setIntegrationAsDefault(integration.id, 'marketing')}
                      disabled={!isOwner}
                    >
                      Use for Marketing
                    </Button>
                  )}

                  {purpose !== 'Transactional Emails' && (
                    <Button
                      onClick={() => setIntegrationAsDefault(integration.id, 'transactional')}
                      disabled={!isOwner}
                    >
                      Use for Transactional
                    </Button>
                  )}
                </Space>
              </div>
            </>
          )}
        </Card>
      )
    })
  }

  // Render provider-specific form fields
  const renderEmailProviderForm = (providerType: EmailProviderKind) => {
    return (
      <>
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
  const renderProviderIcon = (providerKind: EmailProviderKind) => {
    if (providerKind === 'smtp') {
      return <MailOutlined style={{ fontSize: 16 }} />
    } else if (providerKind === 'ses') {
      return <img src="/amazonses.png" alt="Amazon SES" style={{ height: 16 }} />
    } else if (providerKind === 'sparkpost') {
      return <img src="/sparkpost.png" alt="SparkPost" style={{ height: 16 }} />
    } else if (providerKind === 'postmark') {
      return <img src="/postmark.png" alt="Postmark" style={{ height: 16 }} />
    } else if (providerKind === 'mailgun') {
      return <img src="/mailgun.png" alt="Mailgun" style={{ height: 16 }} />
    } else if (providerKind === 'mailjet') {
      return <img src="/mailjet.png" alt="Mailjet" style={{ height: 16 }} />
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

  // Render the email provider selection grid
  const renderEmailProviderGrid = () => {
    return (
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card hoverable onClick={() => handleSelectProviderType('smtp')} className="text-center">
            <MailOutlined style={{ fontSize: 32 }} />
            <p className="mt-2">SMTP</p>
          </Card>
        </Col>
        <Col span={6}>
          <Card hoverable onClick={() => handleSelectProviderType('ses')} className="text-center">
            <img src="/amazonses.png" alt="Amazon SES" style={{ height: 32 }} />
            <p className="mt-2">Amazon SES</p>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            hoverable
            onClick={() => handleSelectProviderType('sparkpost')}
            className="text-center"
          >
            <img src="/sparkpost.png" alt="SparkPost" style={{ height: 32 }} />
            <p className="mt-2">SparkPost</p>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            hoverable
            onClick={() => handleSelectProviderType('mailjet')}
            className="text-center"
          >
            <img src="/mailjet.png" alt="Mailjet" style={{ height: 32 }} />
            <p className="mt-2">Mailjet</p>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            hoverable
            onClick={() => handleSelectProviderType('postmark')}
            className="text-center"
          >
            <img src="/postmark.png" alt="Postmark" style={{ height: 32 }} />
            <p className="mt-2">Postmark</p>
          </Card>
        </Col>
        <Col span={6}>
          <Card
            hoverable
            onClick={() => handleSelectProviderType('mailgun')}
            className="text-center"
          >
            <img src="/mailgun.png" alt="Mailgun" style={{ height: 32 }} />
            <p className="mt-2">Mailgun</p>
          </Card>
        </Col>
      </Row>
    )
  }

  return (
    <Section
      title="Integrations"
      description="Connect and manage external services"
      extra={
        isOwner && !addingEmailProvider && editingIntegrationId === null ? (
          <Button type="primary" onClick={startAddEmailProvider} icon={<PlusOutlined />}>
            Add Integration
          </Button>
        ) : null
      }
    >
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="Email Providers" key="email">
          {addingEmailProvider ? (
            <Card title="Add Email Provider">
              {selectedProviderType ? (
                <Form form={emailProviderForm} layout="vertical" onFinish={saveEmailProvider}>
                  <Form.Item name="kind" hidden initialValue={selectedProviderType}>
                    <Input />
                  </Form.Item>

                  {renderEmailProviderForm(selectedProviderType)}

                  <div className="mt-4 flex justify-end">
                    <Space>
                      <Button onClick={cancelEmailProviderOperation}>Cancel</Button>
                      <Button type="primary" htmlType="submit" loading={loading}>
                        Save
                      </Button>
                    </Space>
                  </div>
                </Form>
              ) : (
                <>
                  <Text>Select an email provider type:</Text>
                  <div className="mt-4">{renderEmailProviderGrid()}</div>
                  <div className="mt-4 flex justify-end">
                    <Button onClick={cancelEmailProviderOperation}>Cancel</Button>
                  </div>
                </>
              )}
            </Card>
          ) : (
            renderEmailIntegrations()
          )}
        </TabPane>
        {/* Additional integration types can be added as tabs in the future */}
      </Tabs>

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
