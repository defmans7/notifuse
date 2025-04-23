import React, { useState } from 'react'
import {
  Card,
  Form,
  Input,
  Select,
  Switch,
  Button,
  Divider,
  Row,
  Col,
  InputNumber,
  Typography,
  Alert
} from 'antd'
import {
  EmailProvider,
  EmailProviderKind,
  AmazonSES,
  SMTPSettings,
  SparkPostSettings,
  Workspace
} from '../services/api/types'
import { MailOutlined, AmazonOutlined, ThunderboltOutlined } from '@ant-design/icons'

const { Option } = Select
const { Title } = Typography

// Form layout constants for consistency
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

export function EmailProviderSettings({
  workspace,
  onSave,
  loading,
  isOwner
}: EmailProviderSettingsProps) {
  const [marketingForm] = Form.useForm()
  const [transactionalForm] = Form.useForm()
  const [marketingProvider, setMarketingProvider] = useState<EmailProviderKind | null>(
    workspace?.settings?.email_marketing?.kind || null
  )
  const [transactionalProvider, setTransactionalProvider] = useState<EmailProviderKind | null>(
    workspace?.settings?.email_transactional?.kind || null
  )

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

  const handleMarketingSave = async (values: any) => {
    if (!workspace) return

    const emailProvider: EmailProvider = {
      kind: values.kind
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

    await onSave(updatedWorkspace)
  }

  const handleTransactionalSave = async (values: any) => {
    if (!workspace) return

    const emailProvider: EmailProvider = {
      kind: values.kind
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
        email_transactional: emailProvider
      }
    }

    await onSave(updatedWorkspace)
  }

  const renderProviderGrid = (
    type: 'marketing' | 'transactional',
    onSelect: (provider: EmailProviderKind) => void
  ) => {
    return (
      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card
            hoverable
            onClick={() => onSelect('smtp')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{ padding: '12px' }}
            size="small"
          >
            <MailOutlined style={{ fontSize: 32, marginBottom: 8 }} />
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 4 }}>
              SMTP
            </Typography.Title>
            <div style={{ fontSize: '12px' }}>Configure with your own SMTP server</div>
          </Card>
        </Col>
        <Col span={8}>
          <Card
            hoverable
            onClick={() => onSelect('ses')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{ padding: '12px' }}
            size="small"
          >
            <AmazonOutlined style={{ fontSize: 32, marginBottom: 8 }} />
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 4 }}>
              Amazon SES
            </Typography.Title>
            <div style={{ fontSize: '12px' }}>Use Amazon Simple Email Service</div>
          </Card>
        </Col>
        <Col span={8}>
          <Card
            hoverable
            onClick={() => onSelect('sparkpost')}
            style={{ textAlign: 'center', height: '100%', padding: '12px' }}
            bodyStyle={{ padding: '12px' }}
            size="small"
          >
            <ThunderboltOutlined style={{ fontSize: 32, marginBottom: 8 }} />
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 4 }}>
              SparkPost
            </Typography.Title>
            <div style={{ fontSize: '12px' }}>Use SparkPost email delivery service</div>
          </Card>
        </Col>
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

    switch (providerType) {
      case 'ses':
        return (
          <Form.Item name={['ses']} initialValue={initialValues?.ses}>
            <Form.Item
              name={['ses', 'region']}
              label="AWS Region"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="us-east-1" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['ses', 'access_key']}
              label="AWS Access Key"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="Access Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['ses', 'secret_key']} label="AWS Secret Key" {...FORM_LAYOUT}>
              <Input.Password placeholder="Secret Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['ses', 'sender_email']}
              label="Sender Email"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="noreply@yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['ses', 'sandbox_mode']}
              valuePropName="checked"
              label="Sandbox Mode"
              {...FORM_LAYOUT}
            >
              <Switch disabled={!isOwner} />
            </Form.Item>
          </Form.Item>
        )

      case 'smtp':
        return (
          <Form.Item name={['smtp']} initialValue={initialValues?.smtp}>
            <Form.Item
              name={['smtp', 'host']}
              label="SMTP Host"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="smtp.yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['smtp', 'port']}
              label="SMTP Port"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <InputNumber min={1} max={65535} placeholder="587" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['smtp', 'username']}
              label="SMTP Username"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="Username" disabled={!isOwner} />
            </Form.Item>
            <Form.Item name={['smtp', 'password']} label="SMTP Password" {...FORM_LAYOUT}>
              <Input.Password placeholder="Password" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['smtp', 'sender_email']}
              label="Sender Email"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="noreply@yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['smtp', 'use_tls']}
              valuePropName="checked"
              label="Use TLS"
              {...FORM_LAYOUT}
            >
              <Switch defaultChecked disabled={!isOwner} />
            </Form.Item>
          </Form.Item>
        )

      case 'sparkpost':
        return (
          <Form.Item name={['sparkpost']} initialValue={initialValues?.sparkpost}>
            <Form.Item name={['sparkpost', 'api_key']} label="SparkPost API Key" {...FORM_LAYOUT}>
              <Input.Password placeholder="API Key" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['sparkpost', 'sender_email']}
              label="Sender Email"
              rules={[{ required: true }]}
              {...FORM_LAYOUT}
            >
              <Input placeholder="noreply@yourdomain.com" disabled={!isOwner} />
            </Form.Item>
            <Form.Item
              name={['sparkpost', 'sandbox_mode']}
              valuePropName="checked"
              label="Sandbox Mode"
              {...FORM_LAYOUT}
            >
              <Switch disabled={!isOwner} />
            </Form.Item>
          </Form.Item>
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
      >
        {!marketingProvider ? (
          <>
            <Alert
              message="An email provider should be connected to send marketing emails"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            {renderProviderGrid('marketing', handleMarketingProviderSelect)}
          </>
        ) : (
          <Form
            form={marketingForm}
            layout="horizontal"
            {...FORM_LAYOUT}
            onFinish={handleMarketingSave}
            initialValues={{
              kind: marketingProvider
            }}
          >
            <Form.Item name="kind" hidden>
              <Input />
            </Form.Item>

            <div style={{ marginBottom: 16, display: 'flex', alignItems: 'center' }}>
              <Title level={4} style={{ marginBottom: 0, marginRight: 'auto' }}>
                {marketingProvider === 'smtp'
                  ? 'SMTP'
                  : marketingProvider === 'ses'
                    ? 'Amazon SES'
                    : 'SparkPost'}{' '}
                Configuration
              </Title>
              <Button
                type="link"
                onClick={() => setMarketingProvider(null)}
                disabled={!isOwner || loading}
              >
                Change provider
              </Button>
            </div>

            {renderProviderForm(marketingProvider, 'marketing')}

            {isOwner && (
              <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  Save Marketing Email Settings
                </Button>
              </Form.Item>
            )}
          </Form>
        )}
      </Card>

      <Card title="Transactional Email Provider" className="workspace-card">
        {!transactionalProvider ? (
          <>
            <Alert
              message="An email provider should be connected to send transactional emails"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />
            {renderProviderGrid('transactional', handleTransactionalProviderSelect)}
          </>
        ) : (
          <Form
            form={transactionalForm}
            layout="horizontal"
            {...FORM_LAYOUT}
            onFinish={handleTransactionalSave}
            initialValues={{
              kind: transactionalProvider
            }}
          >
            <Form.Item name="kind" hidden>
              <Input />
            </Form.Item>

            <div style={{ marginBottom: 16, display: 'flex', alignItems: 'center' }}>
              <Title level={4} style={{ marginBottom: 0, marginRight: 'auto' }}>
                {transactionalProvider === 'smtp'
                  ? 'SMTP'
                  : transactionalProvider === 'ses'
                    ? 'Amazon SES'
                    : 'SparkPost'}{' '}
                Configuration
              </Title>
              <Button
                type="link"
                onClick={() => setTransactionalProvider(null)}
                disabled={!isOwner || loading}
              >
                Change provider
              </Button>
            </div>

            {renderProviderForm(transactionalProvider, 'transactional')}

            {isOwner && (
              <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  Save Transactional Email Settings
                </Button>
              </Form.Item>
            )}
          </Form>
        )}
      </Card>
    </div>
  )
}
