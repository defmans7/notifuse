import { useState } from 'react'
import { Form, Input, Button, Layout, Typography, Card, Tooltip, message } from 'antd'
import { useNavigate } from '@tanstack/react-router'
import { Topbar } from '../components/Topbar'
import { InfoCircleOutlined } from '@ant-design/icons'
import { workspaceService } from '../services/api/workspace'

const { Content } = Layout
const { Title } = Typography

export function CreateWorkspacePage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [form] = Form.useForm()

  // Generate workspace ID from name (alphanumeric only, max 20 chars)
  const generateWorkspaceId = (name: string) => {
    if (!name) return ''
    // Replace spaces with hyphens and remove non-alphanumeric characters
    return name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '')
      .substring(0, 20)
  }

  // Update generated ID when name changes
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const name = e.target.value
    const id = generateWorkspaceId(name)
    form.setFieldsValue({ id })
  }

  const onFinish = async (values: { name: string; id: string; website_url?: string }) => {
    try {
      setLoading(true)
      let logoUrl = null

      // If website URL is provided, detect favicon
      if (values.website_url) {
        try {
          const faviconResponse = await workspaceService.detectFavicon(values.website_url)
          logoUrl = faviconResponse.iconUrl
        } catch (error) {
          console.error('Error detecting favicon:', error)
          // Don't fail the whole process if favicon detection fails
        }
      }

      // Get user's timezone
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone

      // Create workspace with API
      const createResponse = await workspaceService.create({
        id: values.id,
        settings: {
          name: values.name,
          website_url: values.website_url || '',
          logo_url: logoUrl,
          timezone: timezone
        }
      })

      // TODO refresh the auth context workspace list

      message.success(`Workspace "${values.name}" created successfully!`)

      // Navigate to the new workspace
      navigate({
        to: '/workspace/$workspaceId/campaigns',
        params: { workspaceId: values.id }
      })
    } catch (error) {
      console.error('Error creating workspace:', error)
      message.error('Failed to create workspace. Please try again.')
      setLoading(false)
    }
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Topbar />
      <Content style={{ padding: '50px 24px' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: 40 }}>
          Create a New Workspace
        </Title>
        <Card style={{ maxWidth: 600, margin: '0 auto' }}>
          <Form
            name="create-workspace"
            layout="vertical"
            onFinish={onFinish}
            autoComplete="off"
            form={form}
            initialValues={{ id: '' }}
          >
            <Form.Item
              label="Workspace Name"
              name="name"
              rules={[
                { required: true, message: 'Please enter a workspace name' },
                { min: 3, message: 'Workspace name must be at least 3 characters long' }
              ]}
            >
              <Input
                placeholder="Enter a name for your workspace"
                size="large"
                onChange={handleNameChange}
              />
            </Form.Item>

            <Form.Item
              label={
                <span>
                  Workspace ID &nbsp;
                  <Tooltip title="This ID will be used in URLs and API requests. It can only contain lowercase letters, numbers, and hyphens.">
                    <InfoCircleOutlined />
                  </Tooltip>
                </span>
              }
              name="id"
              rules={[
                { required: true, message: 'Workspace ID is required' },
                {
                  pattern: /^[a-z0-9-]+$/,
                  message: 'ID can only contain lowercase letters, numbers, and hyphens'
                }
              ]}
            >
              <Input
                placeholder="workspace-id"
                size="large"
                suffix={
                  <Tooltip title="ID is automatically generated but can be modified if needed">
                    <InfoCircleOutlined style={{ color: 'rgba(0,0,0,.45)' }} />
                  </Tooltip>
                }
              />
            </Form.Item>

            <Form.Item
              label="Website URL"
              name="website_url"
              rules={[
                {
                  pattern: /^(https?:\/\/)?([\da-z.-]+)\.([a-z.]{2,6})([/\w .-]*)*\/?$/,
                  message: 'Please enter a valid URL',
                  validateTrigger: 'onBlur'
                }
              ]}
              extra="We'll automatically detect and use your website's favicon"
            >
              <Input placeholder="https://example.com" size="large" />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                size="large"
                style={{ width: '100%', marginTop: 20 }}
              >
                Create Workspace
              </Button>
            </Form.Item>
          </Form>
        </Card>
      </Content>
    </Layout>
  )
}
