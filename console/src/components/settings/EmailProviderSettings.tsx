import { useState, useEffect, useMemo } from 'react'
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
  Descriptions
} from 'antd'
import {
  EmailProvider,
  EmailProviderKind,
  Workspace,
  Integration,
  CreateIntegrationRequest,
  UpdateIntegrationRequest,
  DeleteIntegrationRequest
} from '../../services/api/types'
import { MailOutlined } from '@ant-design/icons'
import { emailService } from '../../services/api/email'
import { workspaceService } from '../../services/api/workspace'
import { Section } from './Section'

// Constants
const FORM_LAYOUT = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 }
}

// Types
interface EmailProviderSettingsProps {
  workspace: Workspace | null
  onSave: (updatedWorkspace: Workspace) => Promise<void>
  loading: boolean
  isOwner: boolean
}

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

type ProviderType = 'marketing' | 'transactional'

// Utility functions
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

// Sub-components
interface ProviderCardProps {
  provider: EmailProviderKind
  icon: React.ReactNode
  description: string
  onClick: (provider: EmailProviderKind) => void
}

const ProviderCard = ({ provider, icon, onClick }: ProviderCardProps) => (
  <div
    onClick={() => onClick(provider)}
    className="text-center h-full p-3 cursor-pointer hover:bg-gray-50 transition-colors duration-200 flex flex-col items-center justify-center border border-gray-200 rounded-md"
  >
    {icon}
    <div style={{ fontWeight: 'bold', marginTop: '8px' }}>
      {provider === 'smtp'
        ? 'SMTP'
        : provider === 'ses'
          ? 'Amazon SES'
          : provider === 'sparkpost'
            ? 'SparkPost'
            : provider === 'postmark'
              ? 'Postmark'
              : provider === 'mailgun'
                ? 'Mailgun'
                : provider === 'mailjet'
                  ? 'Mailjet'
                  : provider}
    </div>
  </div>
)

interface ProviderGridProps {
  onSelect: (provider: EmailProviderKind) => void
  isTransactional?: boolean
}

const ProviderGrid = ({ onSelect, isTransactional = false }: ProviderGridProps) => {
  const colSpan = isTransactional ? 6 : 8

  return (
    <Row gutter={[16, 16]}>
      <Col span={colSpan}>
        <ProviderCard
          provider="smtp"
          icon={<MailOutlined style={{ fontSize: 24, marginBottom: 8 }} />}
          description="Configure with your own SMTP server"
          onClick={onSelect}
        />
      </Col>
      <Col span={colSpan}>
        <ProviderCard
          provider="ses"
          icon={
            <img src="/amazonses.png" alt="Amazon SES" style={{ height: 24, marginBottom: 8 }} />
          }
          description="Use Amazon Simple Email Service"
          onClick={onSelect}
        />
      </Col>
      <Col span={colSpan}>
        <ProviderCard
          provider="sparkpost"
          icon={
            <img src="/sparkpost.png" alt="SparkPost" style={{ height: 24, marginBottom: 8 }} />
          }
          description="Use SparkPost email delivery service"
          onClick={onSelect}
        />
      </Col>
      <Col span={colSpan}>
        <ProviderCard
          provider="mailjet"
          icon={<img src="/mailjet.png" alt="Mailjet" style={{ height: 24, marginBottom: 8 }} />}
          description="Use Mailjet email delivery service"
          onClick={onSelect}
        />
      </Col>
      {isTransactional && (
        <>
          <Col span={colSpan}>
            <ProviderCard
              provider="postmark"
              icon={
                <img src="/postmark.png" alt="Postmark" style={{ height: 24, marginBottom: 8 }} />
              }
              description="Use Postmark email delivery service"
              onClick={onSelect}
            />
          </Col>
          <Col span={colSpan}>
            <ProviderCard
              provider="mailgun"
              icon={
                <img src="/mailgun.png" alt="Mailgun" style={{ height: 24, marginBottom: 8 }} />
              }
              description="Use Mailgun email delivery service"
              onClick={onSelect}
            />
          </Col>
        </>
      )}
    </Row>
  )
}

interface CommonFormFieldsProps {
  initialValues?: EmailProvider | null
  isOwner: boolean
}

const CommonFormFields = ({ initialValues, isOwner }: CommonFormFieldsProps) => (
  <>
    <Form.Item
      name="default_sender_email"
      label="Sender Email"
      rules={[{ required: true }]}
      initialValue={initialValues?.default_sender_email || ''}
    >
      <Input placeholder="noreply@yourdomain.com" disabled={!isOwner} />
    </Form.Item>
    <Form.Item
      name="default_sender_name"
      label="Sender Name"
      rules={[{ required: true }]}
      initialValue={initialValues?.default_sender_name || 'Notifuse'}
    >
      <Input placeholder="Your Company Name" disabled={!isOwner} />
    </Form.Item>
  </>
)

interface SesFormFieldsProps {
  isOwner: boolean
}

const SesFormFields = ({ isOwner }: SesFormFieldsProps) => (
  <>
    <Form.Item name={['ses', 'region']} label="AWS Region" rules={[{ required: true }]}>
      <Input placeholder="us-east-1" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['ses', 'access_key']} label="AWS Access Key" rules={[{ required: true }]}>
      <Input placeholder="Access Key" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['ses', 'secret_key']} label="AWS Secret Key">
      <Input.Password placeholder="Secret Key" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['ses', 'sandbox_mode']} valuePropName="checked" label="Sandbox Mode">
      <Switch disabled={!isOwner} />
    </Form.Item>
  </>
)

interface SmtpFormFieldsProps {
  isOwner: boolean
}

const SmtpFormFields = ({ isOwner }: SmtpFormFieldsProps) => (
  <>
    <Form.Item name={['smtp', 'host']} label="SMTP Host" rules={[{ required: true }]}>
      <Input placeholder="smtp.yourdomain.com" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['smtp', 'port']} label="SMTP Port" rules={[{ required: true }]}>
      <InputNumber min={1} max={65535} placeholder="587" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['smtp', 'username']} label="SMTP Username" rules={[{ required: true }]}>
      <Input placeholder="Username" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['smtp', 'password']} label="SMTP Password">
      <Input.Password placeholder="Password" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['smtp', 'use_tls']} valuePropName="checked" label="Use TLS">
      <Switch defaultChecked disabled={!isOwner} />
    </Form.Item>
  </>
)

interface SparkpostFormFieldsProps {
  isOwner: boolean
  form: any
}

const SparkpostFormFields = ({ isOwner, form }: SparkpostFormFieldsProps) => (
  <>
    <Form.Item name={['sparkpost', 'endpoint']} label="API Endpoint" rules={[{ required: true }]}>
      <Select
        placeholder="Select SparkPost endpoint"
        disabled={!isOwner}
        options={[
          { label: 'SparkPost US', value: 'https://api.sparkpost.com' },
          { label: 'SparkPost EU', value: 'https://api.eu.sparkpost.com' },
          { label: 'Custom', value: 'custom' }
        ]}
      />
    </Form.Item>

    <Form.Item
      noStyle
      shouldUpdate={(prevValues, currentValues) => {
        return prevValues.sparkpost?.endpoint !== currentValues.sparkpost?.endpoint
      }}
    >
      {({ getFieldValue }) => {
        const endpoint = getFieldValue(['sparkpost', 'endpoint'])
        return endpoint === 'custom' ? (
          <Form.Item
            name={['sparkpost', 'custom_endpoint']}
            label="Custom Endpoint URL"
            rules={[{ required: true }]}
          >
            <Input
              placeholder="https://api.yourdomain.sparkpost.com"
              disabled={!isOwner}
              onChange={(e) => {
                form.setFieldsValue({
                  sparkpost: {
                    ...form.getFieldValue('sparkpost'),
                    endpoint: e.target.value
                  }
                })
              }}
            />
          </Form.Item>
        ) : null
      }}
    </Form.Item>

    <Form.Item name={['sparkpost', 'api_key']} label="SparkPost API Key">
      <Input.Password placeholder="API Key" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['sparkpost', 'sandbox_mode']} valuePropName="checked" label="Sandbox Mode">
      <Switch disabled={!isOwner} />
    </Form.Item>
  </>
)

interface PostmarkFormFieldsProps {
  isOwner: boolean
}

const PostmarkFormFields = ({ isOwner }: PostmarkFormFieldsProps) => (
  <Form.Item name={['postmark', 'server_token']} label="Server Token" rules={[{ required: true }]}>
    <Input.Password placeholder="Server Token" disabled={!isOwner} />
  </Form.Item>
)

interface MailgunFormFieldsProps {
  isOwner: boolean
}

const MailgunFormFields = ({ isOwner }: MailgunFormFieldsProps) => (
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
)

interface MailjetFormFieldsProps {
  isOwner: boolean
}

const MailjetFormFields = ({ isOwner }: MailjetFormFieldsProps) => (
  <>
    <Form.Item name={['mailjet', 'api_key']} label="API Key" rules={[{ required: true }]}>
      <Input.Password placeholder="API Key" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['mailjet', 'secret_key']} label="Secret Key" rules={[{ required: true }]}>
      <Input.Password placeholder="Secret Key" disabled={!isOwner} />
    </Form.Item>
    <Form.Item name={['mailjet', 'sandbox_mode']} valuePropName="checked" label="Sandbox Mode">
      <Switch disabled={!isOwner} />
    </Form.Item>
  </>
)

// Move getProviderConfig and getProviderIntegration to a higher scope as component props
interface ProviderFormProps {
  providerType: EmailProviderKind
  formType: ProviderType
  workspace: Workspace
  form: any
  isOwner: boolean
  getProviderConfig: (type: ProviderType) => EmailProvider | null
}

const ProviderForm = ({
  providerType,
  formType,
  workspace,
  form,
  isOwner,
  getProviderConfig
}: ProviderFormProps) => {
  const initialValues: EmailProvider | null = useMemo(() => {
    if (!workspace) return null

    return getProviderConfig(formType)
  }, [workspace, formType, getProviderConfig])

  if (providerType === 'postmark' && formType === 'marketing') {
    return null // Postmark is only for transactional emails
  }

  if (providerType === 'mailgun' && formType === 'marketing') {
    return null // Mailgun is only for transactional emails
  }

  return (
    <>
      <CommonFormFields initialValues={initialValues} isOwner={isOwner} />

      {providerType === 'ses' && <SesFormFields isOwner={isOwner} />}
      {providerType === 'smtp' && <SmtpFormFields isOwner={isOwner} />}
      {providerType === 'sparkpost' && <SparkpostFormFields isOwner={isOwner} form={form} />}
      {providerType === 'postmark' && <PostmarkFormFields isOwner={isOwner} />}
      {providerType === 'mailgun' && <MailgunFormFields isOwner={isOwner} />}
      {providerType === 'mailjet' && <MailjetFormFields isOwner={isOwner} />}
    </>
  )
}

const getProviderLogo = (providerKind: EmailProviderKind) => {
  if (providerKind === 'smtp') {
    return (
      <Space>
        <MailOutlined style={{ fontSize: 16, marginRight: 8 }} />
        SMTP
      </Space>
    )
  } else if (providerKind === 'ses') {
    return <img src="/amazonses.png" alt="Amazon SES" style={{ height: 16, marginRight: 8 }} />
  } else if (providerKind === 'sparkpost') {
    return <img src="/sparkpost.png" alt="SparkPost" style={{ height: 16, marginRight: 8 }} />
  } else if (providerKind === 'postmark') {
    return <img src="/postmark.png" alt="Postmark" style={{ height: 16, marginRight: 8 }} />
  } else if (providerKind === 'mailgun') {
    return <img src="/mailgun.png" alt="Mailgun" style={{ height: 16, marginRight: 8 }} />
  } else if (providerKind === 'mailjet') {
    return <img src="/mailjet.png" alt="Mailjet" style={{ height: 16, marginRight: 8 }} />
  }
  return null
}

interface ProviderDescriptionProps {
  providerType: ProviderType
  workspace: Workspace
  getProviderConfig: (type: ProviderType) => EmailProvider | null
}

const ProviderDescription = ({
  providerType,
  workspace,
  getProviderConfig
}: ProviderDescriptionProps) => {
  const provider = getProviderConfig(providerType)

  if (!provider) {
    return (
      <Card
        title={`${providerType === 'marketing' ? 'Marketing' : 'Transactional'} Email Provider`}
        className="mb-8"
      >
        <Alert
          message="No email provider configured"
          description={`You have not configured a ${providerType} email provider for this workspace yet.`}
          type="info"
          showIcon
        />
      </Card>
    )
  }

  const logoElement = getProviderLogo(provider.kind)

  const items = [
    {
      key: 'type',
      label: 'Provider Type',
      children: <Space>{logoElement}</Space>
    },
    {
      key: 'sender',
      label: 'Sender Details',
      children: `${provider.default_sender_name} <${provider.default_sender_email}>`
    }
  ]

  // Add provider-specific details
  if (provider.kind === 'smtp' && provider.smtp) {
    items.push(
      {
        key: 'host',
        label: 'SMTP Host',
        children: `${provider.smtp.host}:${provider.smtp.port}`
      },
      {
        key: 'username',
        label: 'SMTP User',
        children: provider.smtp.username
      },
      {
        key: 'security',
        label: 'TLS Enabled',
        children: provider.smtp.use_tls ? 'Yes' : 'No'
      }
    )
  } else if (provider.kind === 'ses' && provider.ses) {
    items.push(
      {
        key: 'region',
        label: 'AWS Region',
        children: provider.ses.region
      },
      {
        key: 'access_key',
        label: 'AWS Access Key',
        children: provider.ses.access_key
      },
      {
        key: 'sandbox',
        label: 'Sandbox Mode',
        children: provider.ses.sandbox_mode ? 'Enabled' : 'Disabled'
      }
    )
  } else if (provider.kind === 'sparkpost' && provider.sparkpost) {
    items.push(
      {
        key: 'endpoint',
        label: 'API Endpoint',
        children: provider.sparkpost.endpoint
      },
      {
        key: 'sandbox',
        label: 'Sandbox Mode',
        children: provider.sparkpost.sandbox_mode ? 'Enabled' : 'Disabled'
      }
    )
  } else if (provider.kind === 'postmark' && provider.postmark) {
    items.push({
      key: 'postmark',
      label: 'Integration',
      children: 'Postmark API Connected'
    })
  } else if (provider.kind === 'mailgun' && provider.mailgun) {
    items.push(
      {
        key: 'domain',
        label: 'Domain',
        children: provider.mailgun.domain
      },
      {
        key: 'region',
        label: 'Region',
        children: provider.mailgun.region || 'US'
      }
    )
  } else if (provider.kind === 'mailjet' && provider.mailjet) {
    items.push(
      {
        key: 'mailjet',
        label: 'Integration',
        children: 'Mailjet API Connected'
      },
      {
        key: 'sandbox',
        label: 'Sandbox Mode',
        children: provider.mailjet.sandbox_mode ? 'Enabled' : 'Disabled'
      }
    )
  }

  return (
    <Descriptions bordered size="small" column={1} style={{ marginBottom: 16 }}>
      {items.map((item) => (
        <Descriptions.Item key={item.key} label={item.label}>
          {item.children}
        </Descriptions.Item>
      ))}
    </Descriptions>
  )
}

interface EmailProviderCardProps {
  providerType: ProviderType
  workspace: Workspace
  provider: EmailProviderKind | null
  editing: boolean
  form: any
  isOwner: boolean
  loading: boolean
  onSelectProvider: (provider: EmailProviderKind) => void
  onEdit: () => void
  onTest: () => void
  onChangeProvider: () => void
  onSave: (values: EmailProviderFormValues) => Promise<void>
  onCancel: () => void
  getProviderConfig: (type: ProviderType) => EmailProvider | null
}

const EmailProviderCard = ({
  providerType,
  workspace,
  provider,
  editing,
  form,
  isOwner,
  loading,
  onSelectProvider,
  onEdit,
  onTest,
  onChangeProvider,
  onSave,
  onCancel,
  getProviderConfig
}: EmailProviderCardProps) => {
  const isTransactional = providerType === 'transactional'
  const title = isTransactional ? 'Transactional Email Provider' : 'Marketing Email Provider'
  const description = isTransactional
    ? 'Used for sending transactional emails like password resets, email verification, and notifications'
    : 'Used for sending marketing emails like newsletters, promotional content, and broadcasts'

  return (
    <Card title={title} className="mb-8 w-full">
      <p className="mb-4 text-gray-600">{description}</p>

      {!provider ? (
        <>
          <Alert
            message="No Email Provider"
            description={`Select an email provider to start sending ${providerType} emails`}
            type="info"
            showIcon
            className="mb-4"
          />
          <div className="mt-4">
            <ProviderGrid onSelect={onSelectProvider} isTransactional={isTransactional} />
          </div>
        </>
      ) : !editing ? (
        <ProviderDescription
          providerType={providerType}
          workspace={workspace}
          getProviderConfig={getProviderConfig}
        />
      ) : (
        <Form
          form={form}
          layout="vertical"
          onFinish={onSave}
          initialValues={getProviderConfig(providerType) || undefined}
        >
          <Form.Item name="kind" hidden>
            <Input />
          </Form.Item>

          <ProviderForm
            providerType={provider}
            formType={providerType}
            workspace={workspace}
            form={form}
            isOwner={isOwner}
            getProviderConfig={getProviderConfig}
          />

          <div className="mt-4 flex justify-end">
            <Space>
              <Button onClick={onCancel}>Cancel</Button>
              <Button type="primary" htmlType="submit" loading={loading} disabled={!isOwner}>
                Save
              </Button>
            </Space>
          </div>
        </Form>
      )}

      {provider && !editing && (
        <div className="mt-4 flex justify-end">
          <Space>
            <Button onClick={onTest} disabled={loading}>
              Test
            </Button>
            {isOwner && (
              <>
                <Button onClick={onEdit} disabled={loading}>
                  Edit
                </Button>
                <Button onClick={onChangeProvider} disabled={loading}>
                  Change Provider
                </Button>
              </>
            )}
          </Space>
        </div>
      )}
    </Card>
  )
}

interface TestEmailModalProps {
  visible: boolean
  loading: boolean
  email: string
  providerType: ProviderType
  onCancel: () => void
  onEmailChange: (email: string) => void
  onSend: () => void
}

const TestEmailModal = ({
  visible,
  loading,
  email,
  providerType,
  onCancel,
  onEmailChange,
  onSend
}: TestEmailModalProps) => (
  <Modal
    title="Test Email Provider"
    open={visible}
    onCancel={onCancel}
    footer={[
      <Button key="cancel" onClick={onCancel}>
        Cancel
      </Button>,
      <Button key="submit" type="primary" loading={loading} onClick={onSend} disabled={!email}>
        Send Test Email
      </Button>
    ]}
  >
    <p>Enter an email address to receive a test email from your {providerType} provider.</p>
    <Input
      placeholder="recipient@example.com"
      value={email}
      onChange={(e) => onEmailChange(e.target.value)}
      style={{ marginBottom: 16 }}
    />
    <Alert
      message="This will send a real test email to the address provided."
      type="info"
      showIcon
    />
  </Modal>
)

// Main component
export function EmailProviderSettings({
  workspace,
  onSave,
  loading,
  isOwner
}: EmailProviderSettingsProps) {
  // Form instances
  const [marketingForm] = Form.useForm()
  const [transactionalForm] = Form.useForm()

  // States for editing mode
  const [editingMarketing, setEditingMarketing] = useState(false)
  const [editingTransactional, setEditingTransactional] = useState(false)

  // States for selected provider type
  const [marketingProvider, setMarketingProvider] = useState<EmailProviderKind | null>(null)
  const [transactionalProvider, setTransactionalProvider] = useState<EmailProviderKind | null>(null)

  // Test email modal state
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmailAddress, setTestEmailAddress] = useState('')
  const [testingProvider, setTestingProvider] = useState<ProviderType | null>(null)
  const [testingEmailLoading, setTestingEmailLoading] = useState(false)

  useEffect(() => {
    if (workspace) {
      // Find and set configured providers
      const marketingProviderIntegration = workspace.settings.marketing_email_provider_id
        ? workspace.integrations?.find(
            (i) => i.id === workspace.settings.marketing_email_provider_id
          )
        : null

      const transactionalProviderIntegration = workspace.settings.transactional_email_provider_id
        ? workspace.integrations?.find(
            (i) => i.id === workspace.settings.transactional_email_provider_id
          )
        : null

      // Set provider types based on integrations
      if (marketingProviderIntegration) {
        setMarketingProvider(marketingProviderIntegration.email_provider.kind)
      }

      if (transactionalProviderIntegration) {
        setTransactionalProvider(transactionalProviderIntegration.email_provider.kind)
      }
    }
  }, [workspace])

  // Get integrations by type
  const getProviderIntegration = (providerType: ProviderType): Integration | null => {
    if (!workspace || !workspace.integrations) return null

    const integrationId =
      providerType === 'marketing'
        ? workspace.settings.marketing_email_provider_id
        : workspace.settings.transactional_email_provider_id

    if (!integrationId) return null

    return workspace.integrations.find((i) => i.id === integrationId) || null
  }

  // Get provider config from integration
  const getProviderConfig = (providerType: ProviderType): EmailProvider | null => {
    const integration = getProviderIntegration(providerType)
    return integration ? integration.email_provider : null
  }

  if (!workspace) {
    return null
  }

  // Handler for selecting a marketing provider
  const handleMarketingProviderSelect = (provider: EmailProviderKind) => {
    setMarketingProvider(provider)
    setEditingMarketing(true)
  }

  // Handler for selecting a transactional provider
  const handleTransactionalProviderSelect = (provider: EmailProviderKind) => {
    setTransactionalProvider(provider)
    setEditingTransactional(true)
  }

  // Handler for canceling marketing provider edit
  const handleMarketingCancel = () => {
    marketingForm.resetFields()
    setEditingMarketing(false)
    // Reset provider if none was previously configured
    if (!workspace?.settings.marketing_email_provider_id) {
      setMarketingProvider(null)
    }
  }

  // Handler for canceling transactional provider edit
  const handleTransactionalCancel = () => {
    transactionalForm.resetFields()
    setEditingTransactional(false)
    // Reset provider if none was previously configured
    if (!workspace?.settings.transactional_email_provider_id) {
      setTransactionalProvider(null)
    }
  }

  // Handler for saving marketing provider
  const handleMarketingSave = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    try {
      const provider = constructProviderFromForm(values)
      const existingIntegration = getProviderIntegration('marketing')

      // If there's an existing integration, update it
      if (existingIntegration) {
        const updateRequest: UpdateIntegrationRequest = {
          workspace_id: workspace.id,
          integration_id: existingIntegration.id,
          name: `Marketing Email Provider (${provider.kind})`,
          provider
        }

        await workspaceService.updateIntegration(updateRequest)
      } else {
        // Create a new integration
        const createRequest: CreateIntegrationRequest = {
          workspace_id: workspace.id,
          name: `Marketing Email Provider (${provider.kind})`,
          type: 'email',
          provider
        }

        const response = await workspaceService.createIntegration(createRequest)

        // Update workspace settings to point to the new integration
        const updatedWorkspace = {
          ...workspace,
          settings: {
            ...workspace.settings,
            marketing_email_provider_id: response.integration_id
          }
        }

        await workspaceService.update({
          id: workspace.id,
          settings: {
            marketing_email_provider_id: response.integration_id
          }
        })

        // Update the local workspace object
        await onSave(updatedWorkspace)
      }

      // Get fresh workspace data with the updated integration
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      setEditingMarketing(false)
      message.success('Marketing email provider saved successfully')
    } catch (error) {
      console.error('Error saving marketing provider', error)
      message.error('Failed to save marketing email provider')
    }
  }

  // Handler for saving transactional provider
  const handleTransactionalSave = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    try {
      const provider = constructProviderFromForm(values)
      const existingIntegration = getProviderIntegration('transactional')

      // If there's an existing integration, update it
      if (existingIntegration) {
        const updateRequest: UpdateIntegrationRequest = {
          workspace_id: workspace.id,
          integration_id: existingIntegration.id,
          name: `Transactional Email Provider (${provider.kind})`,
          provider
        }

        await workspaceService.updateIntegration(updateRequest)
      } else {
        // Create a new integration
        const createRequest: CreateIntegrationRequest = {
          workspace_id: workspace.id,
          name: `Transactional Email Provider (${provider.kind})`,
          type: 'email',
          provider
        }

        const response = await workspaceService.createIntegration(createRequest)

        // Update workspace settings to point to the new integration
        const updatedWorkspace = {
          ...workspace,
          settings: {
            ...workspace.settings,
            transactional_email_provider_id: response.integration_id
          }
        }

        await workspaceService.update({
          id: workspace.id,
          settings: {
            transactional_email_provider_id: response.integration_id
          }
        })

        // Update the local workspace object
        await onSave(updatedWorkspace)
      }

      // Get fresh workspace data with the updated integration
      const response = await workspaceService.get(workspace.id)
      await onSave(response.workspace)

      setEditingTransactional(false)
      message.success('Transactional email provider saved successfully')
    } catch (error) {
      console.error('Error saving transactional provider', error)
      message.error('Failed to save transactional email provider')
    }
  }

  // Handler for opening the test email modal
  const openTestModal = async (provider: ProviderType) => {
    const integration = getProviderIntegration(provider)

    if (!integration) {
      message.error(`No ${provider} email provider configured`)
      return
    }

    setTestEmailAddress('')
    setTestingProvider(provider)
    setTestModalVisible(true)
  }

  // Handler for testing the email provider
  const handleTestProvider = async () => {
    if (!workspace || !testingProvider || !testEmailAddress) return

    try {
      setTestingEmailLoading(true)
      const integration = getProviderIntegration(testingProvider)

      if (!integration) {
        message.error(`No ${testingProvider} email provider configured`)
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

  return (
    <Section title="Email Providers" description="Configure email providers for sending emails">
      <Space className="flex flex-col w-full">
        <EmailProviderCard
          providerType="marketing"
          workspace={workspace}
          provider={marketingProvider}
          editing={editingMarketing}
          form={marketingForm}
          isOwner={isOwner}
          loading={loading}
          onSelectProvider={handleMarketingProviderSelect}
          onEdit={() => setEditingMarketing(true)}
          onTest={() => openTestModal('marketing')}
          onChangeProvider={() => setMarketingProvider(null)}
          onSave={handleMarketingSave}
          onCancel={handleMarketingCancel}
          getProviderConfig={getProviderConfig}
        />

        <EmailProviderCard
          providerType="transactional"
          workspace={workspace}
          provider={transactionalProvider}
          editing={editingTransactional}
          form={transactionalForm}
          isOwner={isOwner}
          loading={loading}
          onSelectProvider={handleTransactionalProviderSelect}
          onEdit={() => setEditingTransactional(true)}
          onTest={() => openTestModal('transactional')}
          onChangeProvider={() => setTransactionalProvider(null)}
          onSave={handleTransactionalSave}
          onCancel={handleTransactionalCancel}
          getProviderConfig={getProviderConfig}
        />

        <TestEmailModal
          visible={testModalVisible}
          loading={testingEmailLoading}
          email={testEmailAddress}
          providerType={testingProvider || 'marketing'}
          onCancel={() => setTestModalVisible(false)}
          onEmailChange={setTestEmailAddress}
          onSend={handleTestProvider}
        />
      </Space>
    </Section>
  )
}
