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
  Descriptions
} from 'antd'
import { EmailProvider, EmailProviderKind, Workspace } from '../services/api/types'
import { MailOutlined } from '@ant-design/icons'
import { emailService } from '../services/api/email'
import { workspaceService } from '../services/api/workspace'

const FORM_LAYOUT = {
  labelCol: { span: 8 },
  wrapperCol: { span: 16 }
}

interface EmailProviderSettingsProps {
  workspace: Workspace | null
  onSave: (updatedWorkspace: Workspace) => Promise<void>
  loading: boolean
  isOwner: boolean
}

// Define the shape of the form values for email provider
interface EmailProviderFormValues {
  kind: EmailProviderKind
  ses?: EmailProvider['ses']
  smtp?: EmailProvider['smtp']
  sparkpost?: EmailProvider['sparkpost']
  postmark?: EmailProvider['postmark']
  default_sender_email: string
  default_sender_name: string
}

export function EmailProviderSettings({
  workspace,
  onSave,
  loading,
  isOwner
}: EmailProviderSettingsProps) {
  const [marketingForm] = Form.useForm<EmailProviderFormValues>()
  const [transactionalForm] = Form.useForm<EmailProviderFormValues>()
  const [marketingProvider, setMarketingProvider] = useState<EmailProviderKind | null>(null)
  const [transactionalProvider, setTransactionalProvider] = useState<EmailProviderKind | null>(null)
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [testLoading, setTestLoading] = useState(false)
  const [providerToTest, setProviderToTest] = useState<'marketing' | 'transactional'>('marketing')
  const [editingMarketing, setEditingMarketing] = useState(false)
  const [editingTransactional, setEditingTransactional] = useState(false)

  // Update provider states when workspace changes
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

    const emailProvider: EmailProvider = {
      kind: values.kind,
      default_sender_email: values.default_sender_email || '',
      default_sender_name: values.default_sender_name || 'Default Sender'
    }

    if (values.kind === 'ses' && values.ses) {
      emailProvider.ses = values.ses
    } else if (values.kind === 'smtp' && values.smtp) {
      emailProvider.smtp = values.smtp
    } else if (values.kind === 'sparkpost' && values.sparkpost) {
      emailProvider.sparkpost = values.sparkpost
    }

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
    } catch (error) {
      message.error('Failed to update marketing email provider')
      console.error(error)
    }
  }

  const handleTransactionalSave = async (values: EmailProviderFormValues) => {
    if (!workspace) return

    const emailProvider: EmailProvider = {
      kind: values.kind,
      default_sender_email: values.default_sender_email || '',
      default_sender_name: values.default_sender_name || 'Default Sender'
    }

    if (values.kind === 'ses' && values.ses) {
      emailProvider.ses = values.ses
    } else if (values.kind === 'smtp' && values.smtp) {
      emailProvider.smtp = values.smtp
    } else if (values.kind === 'sparkpost' && values.sparkpost) {
      emailProvider.sparkpost = values.sparkpost
    } else if (values.kind === 'postmark' && values.postmark) {
      emailProvider.postmark = values.postmark
    }

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
    } catch (error) {
      message.error('Failed to update transactional email provider')
      console.error(error)
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

  // Helper function to construct provider object from form values
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

  const openTestModal = async (provider: 'marketing' | 'transactional') => {
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

  const renderProviderGrid = (
    onSelect: (provider: EmailProviderKind) => void,
    isTransactional = false
  ) => {
    return (
      <Row gutter={[16, 16]}>
        <Col span={isTransactional ? 6 : 8}>
          <Card
            hoverable
            onClick={() => onSelect('smtp')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{
              padding: '12px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center'
            }}
            size="small"
          >
            <MailOutlined style={{ fontSize: 40, marginBottom: 12 }} />
            <div style={{ fontSize: '12px' }}>Configure with your own SMTP server</div>
          </Card>
        </Col>
        <Col span={isTransactional ? 6 : 8}>
          <Card
            hoverable
            onClick={() => onSelect('ses')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{
              padding: '12px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center'
            }}
            size="small"
          >
            <img src="/amazonses.png" alt="Amazon SES" style={{ height: 40, marginBottom: 12 }} />
            <div style={{ fontSize: '12px' }}>Use Amazon Simple Email Service</div>
          </Card>
        </Col>
        <Col span={isTransactional ? 6 : 8}>
          <Card
            hoverable
            onClick={() => onSelect('sparkpost')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{
              padding: '12px',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center'
            }}
            size="small"
          >
            <img src="/sparkpost.png" alt="SparkPost" style={{ height: 40, marginBottom: 12 }} />
            <div style={{ fontSize: '12px' }}>Use SparkPost email delivery service</div>
          </Card>
        </Col>
        {isTransactional && (
          <Col span={6}>
            <Card
              hoverable
              onClick={() => onSelect('postmark')}
              style={{ textAlign: 'center', height: '100%', padding: '12px' }}
              bodyStyle={{
                padding: '12px',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center'
              }}
              size="small"
            >
              <img src="/postmark.png" alt="Postmark" style={{ height: 40, marginBottom: 12 }} />
              <div style={{ fontSize: '12px' }}>Use Postmark email delivery service</div>
            </Card>
          </Col>
        )}
      </Row>
    )
  }

  const renderProviderForm = (
    providerType: EmailProviderKind,
    formType: 'marketing' | 'transactional'
  ) => {
    const initialValues =
      formType === 'marketing'
        ? workspace.settings.email_marketing
        : workspace.settings.email_transactional

    // Common fields for all providers
    const commonFields = (
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

    switch (providerType) {
      case 'ses':
        return (
          <>
            {commonFields}
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
        )

      case 'smtp':
        return (
          <>
            {commonFields}
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
        )

      case 'sparkpost':
        return (
          <>
            {commonFields}
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
                        // When the custom input changes, update the actual endpoint field
                        const form =
                          formType === 'transactional' ? transactionalForm : marketingForm
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
            <Form.Item
              name={['sparkpost', 'sandbox_mode']}
              valuePropName="checked"
              label="Sandbox Mode"
            >
              <Switch disabled={!isOwner} />
            </Form.Item>
          </>
        )

      case 'postmark':
        if (formType === 'marketing') {
          return null // Postmark is only for transactional emails
        }

        return (
          <>
            {commonFields}
            <Form.Item
              name={['postmark', 'server_token']}
              label="Server Token"
              rules={[{ required: true }]}
            >
              <Input.Password placeholder="Server Token" disabled={!isOwner} />
            </Form.Item>
          </>
        )

      default:
        return null
    }
  }

  const renderProviderDescription = (providerType: 'marketing' | 'transactional') => {
    const provider =
      providerType === 'marketing'
        ? workspace.settings.email_marketing
        : workspace.settings.email_transactional

    if (!provider) return null

    // Create provider logo element based on provider kind
    let logoElement
    if (provider.kind === 'smtp') {
      logoElement = (
        <Space>
          <MailOutlined style={{ fontSize: 24, marginRight: 8 }} />
          {provider.kind.toUpperCase()}
        </Space>
      )
    } else if (provider.kind === 'ses') {
      logoElement = (
        <img src="/amazonses.png" alt="Amazon SES" style={{ height: 24, marginRight: 8 }} />
      )
    } else if (provider.kind === 'sparkpost') {
      logoElement = (
        <img src="/sparkpost.png" alt="SparkPost" style={{ height: 24, marginRight: 8 }} />
      )
    } else if (provider.kind === 'postmark') {
      logoElement = (
        <img src="/postmark.png" alt="Postmark" style={{ height: 24, marginRight: 8 }} />
      )
    }

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

  return (
    <div className="email-provider-settings">
      <Card
        title="Marketing Email Provider"
        className="workspace-card"
        style={{ marginBottom: 24 }}
        extra={
          marketingProvider &&
          !editingMarketing && (
            <Space>
              {isOwner && (
                <Button onClick={() => setEditingMarketing(true)} size="small">
                  Edit
                </Button>
              )}
              {isOwner && (
                <Button size="small" onClick={() => openTestModal('marketing')}>
                  Test
                </Button>
              )}
              <Button type="primary" size="small" ghost onClick={() => setMarketingProvider(null)}>
                Change provider
              </Button>
            </Space>
          )
        }
      >
        {!marketingProvider ? (
          <>
            <Alert
              message="An email provider should be connected to send marketing emails"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            {renderProviderGrid(handleMarketingProviderSelect)}
          </>
        ) : !editingMarketing ? (
          renderProviderDescription('marketing')
        ) : (
          <>
            <Form
              form={marketingForm}
              layout="horizontal"
              onFinish={(values) => {
                handleMarketingSave(values).then(() => setEditingMarketing(false))
              }}
              initialValues={{
                kind: marketingProvider
              }}
              {...FORM_LAYOUT}
            >
              <Form.Item name="kind" hidden>
                <Input />
              </Form.Item>

              {renderProviderForm(marketingProvider, 'marketing')}

              {isOwner && (
                <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
                  <Space>
                    <Button onClick={() => setEditingMarketing(false)}>Cancel</Button>
                    <Button onClick={() => openTestModal('marketing')}>Test Integration</Button>
                    <Button type="primary" htmlType="submit" loading={loading}>
                      Save Settings
                    </Button>
                  </Space>
                </Form.Item>
              )}
            </Form>
          </>
        )}
      </Card>

      <Card
        title="Transactional Email Provider"
        className="workspace-card"
        extra={
          transactionalProvider &&
          !editingTransactional && (
            <Space>
              {isOwner && (
                <Button onClick={() => setEditingTransactional(true)} size="small">
                  Edit
                </Button>
              )}
              {isOwner && (
                <Button size="small" onClick={() => openTestModal('transactional')}>
                  Test
                </Button>
              )}
              <Button
                type="primary"
                size="small"
                ghost
                onClick={() => setTransactionalProvider(null)}
              >
                Change provider
              </Button>
            </Space>
          )
        }
      >
        {!transactionalProvider ? (
          <>
            <Alert
              message="An email provider should be connected to send transactional emails"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            {renderProviderGrid(handleTransactionalProviderSelect, true)}
          </>
        ) : !editingTransactional ? (
          renderProviderDescription('transactional')
        ) : (
          <Form
            form={transactionalForm}
            layout="horizontal"
            onFinish={(values) => {
              handleTransactionalSave(values).then(() => setEditingTransactional(false))
            }}
            initialValues={{
              kind: transactionalProvider
            }}
            {...FORM_LAYOUT}
          >
            <Form.Item name="kind" hidden>
              <Input />
            </Form.Item>

            {renderProviderForm(transactionalProvider, 'transactional')}

            {isOwner && (
              <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
                <Space>
                  <Button onClick={() => setEditingTransactional(false)}>Cancel</Button>
                  <Button onClick={() => openTestModal('transactional')}>Test Integration</Button>
                  <Button type="primary" htmlType="submit" loading={loading}>
                    Save Settings
                  </Button>
                </Space>
              </Form.Item>
            )}
          </Form>
        )}
      </Card>

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
            loading={testLoading}
            onClick={handleTestProvider}
            disabled={!testEmail}
          >
            Send Test Email
          </Button>
        ]}
      >
        <p>Enter an email address to receive a test email from your {providerToTest} provider.</p>
        <Input
          placeholder="recipient@example.com"
          value={testEmail}
          onChange={(e) => setTestEmail(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <Alert
          message="This will send a real test email to the address provided."
          type="info"
          showIcon
        />
      </Modal>
    </div>
  )
}
