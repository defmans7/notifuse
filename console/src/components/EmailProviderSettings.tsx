import { useState } from 'react'
import {
  Card,
  Form,
  Input,
  Switch,
  Button,
  Row,
  Col,
  InputNumber,
  Typography,
  Alert,
  Select,
  Modal,
  message,
  Space
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
  const [marketingProvider, setMarketingProvider] = useState<EmailProviderKind | null>(
    workspace?.settings?.email_marketing?.kind || null
  )
  const [transactionalProvider, setTransactionalProvider] = useState<EmailProviderKind | null>(
    workspace?.settings?.email_transactional?.kind || null
  )
  const [testModalVisible, setTestModalVisible] = useState(false)
  const [testEmail, setTestEmail] = useState('')
  const [testLoading, setTestLoading] = useState(false)
  const [providerToTest, setProviderToTest] = useState<'marketing' | 'transactional'>('marketing')

  if (!workspace) {
    return null
  }

  const handleMarketingProviderSelect = (provider: EmailProviderKind) => {
    setMarketingProvider(provider)
    marketingForm.setFieldsValue({ kind: provider })
  }

  const handleTransactionalProviderSelect = (provider: EmailProviderKind) => {
    setTransactionalProvider(provider)
    transactionalForm.setFieldsValue({ kind: provider })
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
      const form = providerToTest === 'marketing' ? marketingForm : transactionalForm
      const formValues = await form.validateFields()

      // Construct provider object from form values
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

  const openTestModal = async (provider: 'marketing' | 'transactional') => {
    const form = provider === 'marketing' ? marketingForm : transactionalForm

    try {
      // Validate the form first
      await form.validateFields()

      // If validation passes, set the provider and show the modal
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

  return (
    <div className="email-provider-settings">
      <Card
        title="Marketing Email Provider"
        className="workspace-card"
        style={{ marginBottom: 24 }}
        extra={
          marketingProvider && (
            <Button type="primary" size="small" ghost onClick={() => setMarketingProvider(null)}>
              Change provider
            </Button>
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
        ) : (
          <Form
            form={marketingForm}
            layout="horizontal"
            onFinish={handleMarketingSave}
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
                  {workspace?.settings?.email_marketing && (
                    <Button onClick={() => openTestModal('marketing')} disabled={loading}>
                      Test Integration
                    </Button>
                  )}
                  <Button type="primary" htmlType="submit" loading={loading}>
                    Save Settings
                  </Button>
                </Space>
              </Form.Item>
            )}
          </Form>
        )}
      </Card>

      <Card
        title="Transactional Email Provider"
        className="workspace-card"
        extra={
          transactionalProvider && (
            <Button
              type="primary"
              size="small"
              ghost
              onClick={() => setTransactionalProvider(null)}
            >
              Change provider
            </Button>
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
        ) : (
          <Form
            form={transactionalForm}
            layout="horizontal"
            onFinish={handleTransactionalSave}
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
                  {workspace?.settings?.email_transactional && (
                    <Button onClick={() => openTestModal('transactional')} disabled={loading}>
                      Test Integration
                    </Button>
                  )}
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={loading}
                    style={{ marginRight: 8 }}
                  >
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
