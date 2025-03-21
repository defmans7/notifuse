import { useState } from 'react'
import {
  Card,
  Typography,
  Space,
  Descriptions,
  Button,
  Modal,
  Form,
  Input,
  Select,
  App
} from 'antd'
import { SettingOutlined, EditOutlined } from '@ant-design/icons'
import { Workspace } from '../services/api/types'
import { workspaceService } from '../services/api/workspace'

const { Option } = Select

interface WorkspaceSettingsProps {
  workspace: Workspace | null
  loading: boolean
  onWorkspaceUpdate: (workspace: Workspace) => void
}

export function WorkspaceSettings({
  workspace,
  loading,
  onWorkspaceUpdate
}: WorkspaceSettingsProps) {
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [form] = Form.useForm()
  const { message } = App.useApp()

  const showEditModal = () => {
    // Set form values from workspace data
    form.setFieldsValue({
      name: workspace?.name || '',
      website_url: workspace?.settings.website_url || '',
      timezone: workspace?.settings.timezone || 'UTC'
    })
    setEditModalVisible(true)
  }

  const handleSaveSettings = async (values: any) => {
    if (!workspace) return

    setSavingSettings(true)
    try {
      await workspaceService.update({
        id: workspace.id,
        name: values.name,
        settings: {
          website_url: values.website_url,
          logo_url: workspace?.settings.logo_url || null,
          cover_url: workspace?.settings.cover_url || null,
          timezone: values.timezone
        }
      })

      // Refresh the workspace data
      const response = await workspaceService.get(workspace.id)

      // Update the parent component with the new workspace data
      onWorkspaceUpdate(response.workspace)

      // Close the modal and show success message
      setEditModalVisible(false)
      message.success('Workspace settings updated successfully')
    } catch (error) {
      console.error('Failed to update workspace settings', error)
      message.error('Failed to update workspace settings')
    } finally {
      setSavingSettings(false)
    }
  }

  return (
    <>
      <Card
        title={
          <Space>
            <SettingOutlined />
            <span>General Settings</span>
          </Space>
        }
        extra={
          <Button icon={<EditOutlined />} onClick={showEditModal} disabled={loading || !workspace}>
            Edit
          </Button>
        }
        loading={loading}
      >
        {workspace && (
          <Descriptions column={1} bordered>
            <Descriptions.Item label="Workspace Name">{workspace.name}</Descriptions.Item>
            <Descriptions.Item label="Website URL">
              {workspace.settings.website_url || 'Not specified'}
            </Descriptions.Item>
            <Descriptions.Item label="Timezone">{workspace.settings.timezone}</Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Modal
        title="Edit Workspace Settings"
        open={editModalVisible}
        onCancel={() => setEditModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleSaveSettings}>
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
            <Space>
              <Button onClick={() => setEditModalVisible(false)}>Cancel</Button>
              <Button type="primary" htmlType="submit" loading={savingSettings}>
                Save Changes
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </>
  )
}
