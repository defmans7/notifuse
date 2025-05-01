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
import { EmailProvider, EmailProviderKind, Workspace } from '../services/api/types'
import { MailOutlined } from '@ant-design/icons'
import { emailService } from '../services/api/email'
import { workspaceService } from '../services/api/workspace'

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

const ProviderCard = ({ provider, icon, description, onClick }: ProviderCardProps) => (
  <Card
    hoverable
    onClick={() => onClick(provider)}
    style={{ textAlign: 'center', height: '100%', padding: '12px' }}
    styles={{
      body: {
        padding: '12px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center'
      }
    }}
    size="small"
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
              : provider}
    </div>
  </Card>
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
      {isTransactional && (
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

interface ProviderFormProps {
  providerType: EmailProviderKind
  formType: ProviderType
  workspace: Workspace
  form: any
  isOwner: boolean
}

const ProviderForm = ({ providerType, formType, workspace, form, isOwner }: ProviderFormProps) => {
  const initialValues =
    formType === 'marketing'
      ? workspace.settings.email_marketing
      : workspace.settings.email_transactional

  if (providerType === 'postmark' && formType === 'marketing') {
    return null // Postmark is only for transactional emails
  }

  return (
    <>
      <CommonFormFields initialValues={initialValues} isOwner={isOwner} />

      {providerType === 'ses' && <SesFormFields isOwner={isOwner} />}
      {providerType === 'smtp' && <SmtpFormFields isOwner={isOwner} />}
      {providerType === 'sparkpost' && <SparkpostFormFields isOwner={isOwner} form={form} />}
      {providerType === 'postmark' && <PostmarkFormFields isOwner={isOwner} />}
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
  }
  return null
}

interface ProviderDescriptionProps {
  providerType: ProviderType
  workspace: Workspace
}

const ProviderDescription = ({ providerType, workspace }: ProviderDescriptionProps) => {
  const provider =
    providerType === 'marketing'
      ? workspace.settings.email_marketing
      : workspace.settings.email_transactional

  if (!provider) return null

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
  title: string
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
}

const EmailProviderCard = ({
  title,
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
  onCancel
}: EmailProviderCardProps) => {
  const isTransactional = providerType === 'transactional'

  return (
    <Card
      title={title}
      className="workspace-card"
      style={{ marginBottom: isTransactional ? 0 : 24 }}
      extra={
        provider &&
        !editing && (
          <Space>
            {isOwner && (
              <Button onClick={onEdit} size="small">
                Edit
              </Button>
            )}
            {isOwner && (
              <Button size="small" onClick={onTest}>
                Test
              </Button>
            )}
            <Button type="primary" size="small" ghost onClick={onChangeProvider}>
              Change provider
            </Button>
          </Space>
        )
      }
    >
      {!provider ? (
        <>
          <Alert
            message={`An email provider should be connected to send ${providerType} emails`}
            type="info"
            showIcon
            style={{ marginBottom: 16 }}
          />
          <ProviderGrid onSelect={onSelectProvider} isTransactional={isTransactional} />
        </>
      ) : !editing ? (
        <ProviderDescription providerType={providerType} workspace={workspace} />
      ) : (
        <Form
          form={form}
          layout="horizontal"
          onFinish={(values) => onSave(values)}
          initialValues={{ kind: provider }}
          {...FORM_LAYOUT}
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
          />

          {isOwner && (
            <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
              <Space>
                <Button onClick={onCancel}>Cancel</Button>
                <Button onClick={onTest}>Test Integration</Button>
                <Button type="primary" htmlType="submit" loading={loading}>
                  Save Settings
                </Button>
              </Space>
            </Form.Item>
          )}
        </Form>
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
  const [marketingForm] = Form.useForm<EmailProviderFormValues>()
  const [transactionalForm] = Form.useForm<EmailProviderFormValues>()

  // Provider state
  const [marketingProvider, setMarketingProvider] = useState<EmailProviderKind | null>(null)
  const [transactionalProvider, setTransactionalProvider] = useState<EmailProviderKind | null>(null)

  // Editing state
  const [editingMarketing, setEditingMarketing] = useState(false)
  const [editingTransactional, setEditingTransactional] = useState(false)

  // Test email modal state
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [testLoading, setTestLoading] = useState(false)
  const [providerToTest, setProviderToTest] = useState<ProviderType>('marketing')

  // Initialize forms from workspace settings
  useEffect(() => {
    if (workspace) {
      // Update marketing provider
      if (workspace.settings?.email_marketing?.kind) {
        setMarketingProvider(workspace.settings.email_marketing.kind)
        marketingForm.setFieldsValue({
          ...workspace.settings.email_marketing,
          kind: workspace.settings.email_marketing.kind
        })
      }

      // Update transactional provider
      if (workspace.settings?.email_transactional?.kind) {
        setTransactionalProvider(workspace.settings.email_transactional.kind)
        transactionalForm.setFieldsValue({
          ...workspace.settings.email_transactional,
          kind: workspace.settings.email_transactional.kind
        })
      }
    }
  }, [workspace, marketingForm, transactionalForm])

  if (!workspace) {
    return null
  }

  // Handler functions
  const handleMarketingProviderSelect = (provider: EmailProviderKind) => {
    setMarketingProvider(provider)
    marketingForm.setFieldsValue({ kind: provider })
    setEditingMarketing(true)
  }

  const handleTransactionalProviderSelect = (provider: EmailProviderKind) => {
    setTransactionalProvider(provider)
    transactionalForm.setFieldsValue({ kind: provider })
    setEditingTransactional(true)
  }

  const handleMarketingSave = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    const emailProvider = constructProviderFromForm(values)

    const updatedWorkspace = {
      ...workspace,
      settings: {
        ...workspace.settings,
        email_marketing: emailProvider
      }
    }

    try {
      const response = await workspaceService.update(updatedWorkspace)
      message.success('Marketing email provider updated successfully')
      await onSave(response.workspace)
      setEditingMarketing(false)
    } catch (error) {
      message.error('Failed to update marketing email provider')
      console.error(error)
    }
  }

  const handleTransactionalSave = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    const emailProvider = constructProviderFromForm(values)

    const updatedWorkspace = {
      ...workspace,
      settings: {
        ...workspace.settings,
        email_transactional: emailProvider
      }
    }

    try {
      const response = await workspaceService.update(updatedWorkspace)
      message.success('Transactional email provider updated successfully')
      await onSave(response.workspace)
      setEditingTransactional(false)
    } catch (error) {
      message.error('Failed to update transactional email provider')
      console.error(error)
    }
  }

  const openTestModal = async (provider: ProviderType) => {
    // Only validate the form if we're in editing mode
    const isEditing = provider === 'marketing' ? editingMarketing : editingTransactional
    const form = provider === 'marketing' ? marketingForm : transactionalForm

    // Check if there's a configured provider when not editing
    if (!isEditing) {
      const existingProvider =
        provider === 'marketing'
          ? workspace?.settings?.email_marketing
          : workspace?.settings?.email_transactional

      if (!existingProvider) {
        message.error('No email provider configured')
        return
      }

      // If not editing and provider exists, show test modal directly
      setProviderToTest(provider)
      setTestModalVisible(true)
      return
    }

    try {
      // Validate the form when in editing mode
      await form.validateFields()
      setProviderToTest(provider)
      setTestModalVisible(true)
    } catch (error) {
      // If validation fails, show error message
      message.error('Please fill all required fields correctly before testing')
    }
  }

  const handleTestProvider = async () => {
    if (!workspace || !testEmail) return

    setTestLoading(true)
    try {
      let provider: EmailProvider

      if (providerToTest === 'marketing') {
        if (editingMarketing) {
          // Use form values when editing
          const formValues = await marketingForm.validateFields()
          provider = constructProviderFromForm(formValues)
        } else {
          // Use current workspace settings when not editing
          if (!workspace.settings.email_marketing) {
            throw new Error('No marketing email provider configured')
          }
          provider = workspace.settings.email_marketing
        }
      } else {
        // transactional
        if (editingTransactional) {
          // Use form values when editing
          const formValues = await transactionalForm.validateFields()
          provider = constructProviderFromForm(formValues)
        } else {
          // Use current workspace settings when not editing
          if (!workspace.settings.email_transactional) {
            throw new Error('No transactional email provider configured')
          }
          provider = workspace.settings.email_transactional
        }
      }

      const response = await emailService.testProvider(workspace.id, provider, testEmail)

      if (response.success) {
        message.success('Test email sent successfully')
        setTestModalVisible(false)
        setTestEmail('')
      } else {
        message.error(`Failed to send test email: ${response.error || 'Unknown error'}`)
      }
    } catch (error) {
      if (error instanceof Error) {
        message.error(`Error: ${error.message}`)
      } else {
        message.error('Failed to validate form')
      }
    } finally {
      setTestLoading(false)
    }
  }

  return (
    <div className="email-provider-settings">
      <EmailProviderCard
        title="Marketing Email Provider"
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
        onCancel={() => setEditingMarketing(false)}
      />

      <EmailProviderCard
        title="Transactional Email Provider"
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
        onCancel={() => setEditingTransactional(false)}
      />

      <TestEmailModal
        visible={testModalVisible}
        loading={testLoading}
        email={testEmail}
        providerType={providerToTest}
        onCancel={() => setTestModalVisible(false)}
        onEmailChange={setTestEmail}
        onSend={handleTestProvider}
      />
    </div>
  )
}
