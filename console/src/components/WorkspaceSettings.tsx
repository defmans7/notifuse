import { useState } from 'react'
import {
  Card,
  Space,
  Descriptions,
  Button,
  Modal,
  Form,
  Input,
  Select,
  App,
  Typography,
  Divider,
  Tooltip
} from 'antd'
import { EditOutlined, DeleteOutlined, ExclamationCircleOutlined } from '@ant-design/icons'
import { Workspace } from '../services/api/types'
import { workspaceService } from '../services/api/workspace'
import { useNavigate } from '@tanstack/react-router'

const { Option } = Select
const { Text } = Typography

interface WorkspaceSettingsProps {
  workspace: Workspace | null
  loading: boolean
  onWorkspaceUpdate: (workspace: Workspace) => void
  onWorkspaceDelete?: () => void
  isOwner: boolean
}

export function WorkspaceSettings({
  workspace,
  loading,
  onWorkspaceUpdate,
  onWorkspaceDelete,
  isOwner
}: WorkspaceSettingsProps) {
  const [editModalVisible, setEditModalVisible] = useState(false)
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [deletingWorkspace, setDeletingWorkspace] = useState(false)
  const [confirmWorkspaceId, setConfirmWorkspaceId] = useState('')
  const [form] = Form.useForm()
  const { message } = App.useApp()
  const navigate = useNavigate()

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

  const showDeleteModal = () => {
    setDeleteModalVisible(true)
    setConfirmWorkspaceId('')
  }

  const handleDeleteWorkspace = async () => {
    if (!workspace) return

    if (confirmWorkspaceId !== workspace.id) {
      message.error('Workspace ID does not match')
      return
    }

    setDeletingWorkspace(true)
    try {
      await workspaceService.delete({ id: workspace.id })
      message.success('Workspace deleted successfully')
      // Call parent callback to refresh workspaces
      if (onWorkspaceDelete) {
        onWorkspaceDelete()
      } else {
        // Fallback if no callback provided
        navigate({ to: '/' })
      }
    } catch (error) {
      console.error('Failed to delete workspace', error)
      message.error('Failed to delete workspace')
    } finally {
      setDeletingWorkspace(false)
      setDeleteModalVisible(false)
    }
  }

  return (
    <>
      <Card
        title="General Settings"
        extra={
          <Space>
            {workspace && isOwner && (
              <>
                <Tooltip title="Delete workspace">
                  <Button
                    danger
                    type="text"
                    ghost
                    size="small"
                    onClick={showDeleteModal}
                    icon={<DeleteOutlined />}
                  />
                </Tooltip>
                <Button
                  type="primary"
                  size="small"
                  ghost
                  onClick={showEditModal}
                  disabled={loading || !workspace}
                >
                  <EditOutlined /> Edit
                </Button>
              </>
            )}
          </Space>
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

      {/* Edit Workspace Modal */}
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

      {/* Delete Workspace Confirmation Modal */}
      <Modal
        title={
          <span>
            <ExclamationCircleOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
            Delete Workspace
          </span>
        }
        open={deleteModalVisible}
        onCancel={() => setDeleteModalVisible(false)}
        footer={null}
      >
        <div>
          <Text strong>Warning: This action cannot be undone.</Text>
          <p style={{ marginTop: 16 }}>
            This will permanently delete the workspace "{workspace?.name}" and all of its data.
          </p>
          <p>
            To confirm, please enter the workspace ID: <Text code>{workspace?.id}</Text>
          </p>

          <Input
            value={confirmWorkspaceId}
            onChange={(e) => setConfirmWorkspaceId(e.target.value)}
            placeholder="Enter workspace ID"
            style={{ marginBottom: 16 }}
          />

          <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 24 }}>
            <Space>
              <Button onClick={() => setDeleteModalVisible(false)}>Cancel</Button>
              <Button
                danger
                type="primary"
                loading={deletingWorkspace}
                disabled={confirmWorkspaceId !== workspace?.id}
                onClick={handleDeleteWorkspace}
              >
                Delete Workspace
              </Button>
            </Space>
          </div>
        </div>
      </Modal>
    </>
  )
}
