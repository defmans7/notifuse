import { Card, Form, Input, Button } from 'antd'
import { useNavigate } from '@tanstack/react-router'
import { useAuth } from '../../contexts/AuthContext'
import config from '../../config'
import { useState } from 'react'
import { message } from 'antd'
import { createFileRoute } from '@tanstack/react-router'

interface WorkspaceForm {
  name: string
  id: string
  url: string
}

function toSnakeCase(str: string): string {
  return str
    .toLowerCase()
    .replace(/\s+/g, '_') // Replace spaces with underscores
    .replace(/[^a-z0-9_]/g, '') // Remove any character that's not alphanumeric or underscore
}

function CreateWorkspace() {
  const { refreshWorkspaces } = useAuth()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [generatedId, setGeneratedId] = useState('')

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const snakeCaseId = toSnakeCase(e.target.value)
    setGeneratedId(snakeCaseId)
    form.setFieldValue('id', snakeCaseId)
  }

  const handleSubmit = async (values: WorkspaceForm) => {
    try {
      const authToken = localStorage.getItem('auth_token')
      if (!authToken) {
        throw new Error('No authentication token')
      }

      const response = await fetch(`${config.API_ENDPOINT}/workspaces.create`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`
        },
        body: JSON.stringify({
          id: values.id,
          settings: {
            name: values.name,
            url: values.url,
            logo_url: null,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
          }
        })
      })

      if (!response.ok) {
        throw new Error('Failed to create workspace')
      }

      const data = await response.json()
      const workspaceId = data.workspace.id

      await refreshWorkspaces()
      navigate({
        to: '/workspace/$workspaceId',
        params: { workspaceId }
      })
      message.success('Workspace created successfully')
    } catch (error) {
      message.error('Failed to create workspace')
    }
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto', padding: '24px' }}>
      <Card title="Create Your First Workspace">
        <p style={{ marginBottom: 24 }}>
          A workspace is where you'll manage your email templates and campaigns.
        </p>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          validateMessages={{
            required: '${label} is required'
          }}
        >
          <Form.Item
            label="Workspace Name"
            name="name"
            rules={[{ required: true }]}
            tooltip="This will be displayed in the navigation and workspace switcher"
          >
            <Input placeholder="My Company" onChange={handleNameChange} />
          </Form.Item>

          <Form.Item
            label="Workspace ID"
            name="id"
            rules={[
              { required: true },
              {
                pattern: /^[a-z0-9_]+$/,
                message: 'ID can only contain lowercase letters, numbers, and underscores'
              }
            ]}
            tooltip="A unique identifier for your workspace. This will be used in URLs and API calls."
          >
            <Input placeholder="my_company" value={generatedId} />
          </Form.Item>

          <Form.Item
            label="Website URL"
            name="url"
            rules={[{ required: true }, { type: 'url', message: 'Please enter a valid URL' }]}
            tooltip="The website associated with this workspace"
          >
            <Input placeholder="https://example.com" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              Create Workspace
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export const Route = createFileRoute('/workspace/create')({
  component: CreateWorkspace
})
