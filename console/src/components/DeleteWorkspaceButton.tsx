import { useState } from 'react'
import { Button, Modal, Input, Space, Typography, App } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'
import { Workspace } from '../services/api/types'
import { workspaceService } from '../services/api/workspace'

const { Text } = Typography

interface DeleteWorkspaceButtonProps {
  workspace: Workspace
  onDeleteSuccess: () => void
}

export function DeleteWorkspaceButton({ workspace, onDeleteSuccess }: DeleteWorkspaceButtonProps) {
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [deletingWorkspace, setDeletingWorkspace] = useState(false)
  const [confirmWorkspaceId, setConfirmWorkspaceId] = useState('')
  const { message } = App.useApp()

  const showDeleteModal = () => {
    setDeleteModalVisible(true)
    setConfirmWorkspaceId('')
  }

  const handleDeleteWorkspace = async () => {
    if (confirmWorkspaceId !== workspace.id) {
      message.error('Workspace ID does not match')
      return
    }

    setDeletingWorkspace(true)
    try {
      await workspaceService.delete({ id: workspace.id })
      message.success('Workspace deleted successfully')
      onDeleteSuccess()
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
      <Button type="default" size="small" onClick={showDeleteModal}>
        Delete
      </Button>

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
            This will permanently delete the workspace "{workspace.name}" and all of its data.
          </p>
          <p>
            To confirm, please enter the workspace ID: <Text code>{workspace.id}</Text>
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
                disabled={confirmWorkspaceId !== workspace.id}
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
