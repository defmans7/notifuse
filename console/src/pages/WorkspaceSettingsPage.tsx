import { useState, useEffect } from 'react'
import { Card, Typography, Divider, Space, Spin, Form, Input, Select, Button, App } from 'antd'
import { useParams } from '@tanstack/react-router'
import { workspaceService } from '../services/api/workspace'
import { Workspace } from '../services/api/types'
import { SettingOutlined } from '@ant-design/icons'
import { WorkspaceMembers } from '../components/WorkspaceMembers'

const { Title } = Typography
const { Option } = Select

export function WorkspaceSettingsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/settings' })
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loadingWorkspace, setLoadingWorkspace] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  useEffect(() => {
    async function fetchWorkspace() {
      setLoadingWorkspace(true)
      try {
        const response = await workspaceService.get(workspaceId)
        setWorkspace(response.workspace)
        form.setFieldsValue({
          name: response.workspace.name,
          website_url: response.workspace.settings.website_url,
          timezone: response.workspace.settings.timezone
        })
      } catch (error) {
        console.error('Failed to fetch workspace', error)
      } finally {
        setLoadingWorkspace(false)
      }
    }

    fetchWorkspace()
  }, [workspaceId, form])

  const handleSaveSettings = async (values: any) => {
    setSavingSettings(true)
    try {
      await workspaceService.update({
        id: workspaceId,
        name: values.name,
        settings: {
          website_url: values.website_url,
          logo_url: workspace?.settings.logo_url || null,
          cover_url: workspace?.settings.cover_url || null,
          timezone: values.timezone
        }
      })

      // Refresh the workspace data
      const response = await workspaceService.get(workspaceId)
      setWorkspace(response.workspace)

      // Show success message
      message.success('Workspace settings updated successfully')
    } catch (error) {
      console.error('Failed to update workspace settings', error)
      message.error('Failed to update workspace settings')
    } finally {
      setSavingSettings(false)
    }
  }

  return (
    <div className="workspace-settings">
      <Title level={2}>Workspace Settings</Title>
      <Divider />

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card
          title={
            <Space>
              <SettingOutlined />
              <span>General Settings</span>
            </Space>
          }
          loading={loadingWorkspace}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSaveSettings}
            initialValues={{
              name: workspace?.name || '',
              website_url: workspace?.settings.website_url || '',
              timezone: workspace?.settings.timezone || 'UTC'
            }}
          >
            <Form.Item
              name="name"
              label="Workspace Name"
              rules={[{ required: true, message: 'Please enter workspace name' }]}
            >
              <Input placeholder="Enter workspace name" />
            </Form.Item>

            <Form.Item name="website_url" label="Website URL">
              <Input placeholder="https://example.com" />
            </Form.Item>

            <Form.Item
              name="timezone"
              label="Timezone"
              rules={[{ required: true, message: 'Please select a timezone' }]}
            >
              <Select>
                <Option value="UTC">UTC</Option>
                <Option value="America/New_York">Eastern Time (ET)</Option>
                <Option value="America/Chicago">Central Time (CT)</Option>
                <Option value="America/Denver">Mountain Time (MT)</Option>
                <Option value="America/Los_Angeles">Pacific Time (PT)</Option>
                <Option value="Europe/London">London</Option>
                <Option value="Europe/Paris">Paris</Option>
                <Option value="Asia/Tokyo">Tokyo</Option>
              </Select>
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={savingSettings}>
                Save Changes
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <WorkspaceMembers workspaceId={workspaceId} />
      </Space>
    </div>
  )
}
