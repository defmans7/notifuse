import { useState, useEffect } from 'react'
import { Typography, Divider, Space } from 'antd'
import { useParams } from '@tanstack/react-router'
import { workspaceService } from '../services/api/workspace'
import { Workspace } from '../services/api/types'
import { WorkspaceMembers } from '../components/WorkspaceMembers'
import { WorkspaceSettings } from '../components/WorkspaceSettings'

const { Title } = Typography

export function WorkspaceSettingsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/settings' })
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loadingWorkspace, setLoadingWorkspace] = useState(false)

  useEffect(() => {
    fetchWorkspace()
  }, [workspaceId])

  const fetchWorkspace = async () => {
    setLoadingWorkspace(true)
    try {
      const response = await workspaceService.get(workspaceId)
      setWorkspace(response.workspace)
    } catch (error) {
      console.error('Failed to fetch workspace', error)
    } finally {
      setLoadingWorkspace(false)
    }
  }

  const handleWorkspaceUpdate = (updatedWorkspace: Workspace) => {
    setWorkspace(updatedWorkspace)
  }

  return (
    <div className="workspace-settings">
      <Title level={2}>Workspace Settings</Title>
      <Divider />

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <WorkspaceSettings
          workspace={workspace}
          loading={loadingWorkspace}
          onWorkspaceUpdate={handleWorkspaceUpdate}
        />

        <WorkspaceMembers workspaceId={workspaceId} />
      </Space>
    </div>
  )
}
