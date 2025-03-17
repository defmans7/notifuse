import { Card, Typography, Button, Modal, Input, message } from 'antd'
import { useParams, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useAuth } from '../../../contexts/AuthContext'
import config from '../../../config'
import { DeleteOutlined } from '@ant-design/icons'
import { createFileRoute } from '@tanstack/react-router'

const { Title, Text } = Typography

function WorkspaceSettings() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId' })
  const { workspaces, refreshWorkspaces } = useAuth()
  const navigate = useNavigate()
  const [isDeleteModalVisible, setIsDeleteModalVisible] = useState(false)
  const [confirmWorkspaceId, setConfirmWorkspaceId] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)

  const workspace = workspaces.find((w) => w.id === workspaceId)

  if (!workspace) {
    return null
  }

  const handleDelete = async () => {
    if (confirmWorkspaceId !== workspace.id) {
      message.error('Workspace ID does not match')
      return
    }

    try {
      setIsDeleting(true)
      const authToken = localStorage.getItem('auth_token')
      if (!authToken) {
        throw new Error('No authentication token')
      }

      const response = await fetch(`${config.API_ENDPOINT}/workspaces.delete`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${authToken}`
        },
        body: JSON.stringify({
          workspace_id: workspace.id
        })
      })

      if (!response.ok) {
        throw new Error('Failed to delete workspace')
      }

      await refreshWorkspaces()
      message.success('Workspace deleted successfully')
      navigate({ to: '/' })
    } catch (error) {
      message.error('Failed to delete workspace')
    } finally {
      setIsDeleting(false)
      setIsDeleteModalVisible(false)
    }
  }

  return (
    <div>
      <Title level={2}>Workspace Settings</Title>

      <Card
        title="Danger Zone"
        style={{ marginTop: 24 }}
        headStyle={{ backgroundColor: '#fff1f0', borderBottom: '1px solid #ffa39e' }}
        bodyStyle={{ backgroundColor: '#fff2f0' }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <Text strong>Delete Workspace</Text>
            <br />
            <Text type="secondary">
              Once you delete a workspace, there is no going back. Please be certain.
            </Text>
          </div>
          <Button
            danger
            type="primary"
            icon={<DeleteOutlined />}
            onClick={() => setIsDeleteModalVisible(true)}
          >
            Delete Workspace
          </Button>
        </div>
      </Card>

      <Modal
        title="Delete Workspace"
        open={isDeleteModalVisible}
        onCancel={() => setIsDeleteModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setIsDeleteModalVisible(false)}>
            Cancel
          </Button>,
          <Button
            key="delete"
            danger
            type="primary"
            loading={isDeleting}
            disabled={confirmWorkspaceId !== workspace.id}
            onClick={handleDelete}
          >
            Delete Workspace
          </Button>
        ]}
      >
        <p>
          This action cannot be undone. This will permanently delete the{' '}
          <strong>{workspace.settings.name}</strong> workspace and all of its data.
        </p>
        <p>
          Please type <strong>{workspace.id}</strong> to confirm.
        </p>
        <Input
          placeholder="Enter workspace ID"
          value={confirmWorkspaceId}
          onChange={(e) => setConfirmWorkspaceId(e.target.value)}
          style={{ marginTop: 12 }}
        />
      </Modal>
    </div>
  )
}

export const Route = createFileRoute('/workspace/$workspaceId/settings')({
  component: WorkspaceSettings
})
