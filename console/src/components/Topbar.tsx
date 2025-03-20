import { Layout, Select, Space } from 'antd'
import { LogoutOutlined } from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'
import logo from '../assets/logo.png'
import { useNavigate, useMatch } from '@tanstack/react-router'
import { Workspace } from '../services/api/types'

const { Header } = Layout

export function Topbar() {
  const { signout, workspaces } = useAuth()
  const navigate = useNavigate()
  const workspaceMatch = useMatch({ from: '/workspace/$workspaceId', shouldThrow: false })
  const currentWorkspaceId = workspaceMatch?.params?.workspaceId

  const handleWorkspaceChange = (workspaceId: string) => {
    navigate({
      to: '/workspace/$workspaceId/campaigns',
      params: { workspaceId }
    })
  }

  return (
    <Header
      style={{
        padding: '0 24px',
        background: '#fff',
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        boxShadow: '0 2px 8px rgba(0, 0, 0, 0.06)'
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center' }}>
        <img
          src={logo}
          alt="Logo"
          style={{ height: '32px', cursor: 'pointer' }}
          onClick={() => navigate({ to: '/' })}
        />

        {workspaces.length > 0 && (
          <Select
            value={currentWorkspaceId}
            onChange={handleWorkspaceChange}
            style={{ width: 200, marginLeft: 24 }}
            placeholder="Select workspace"
            options={workspaces.map((workspace: Workspace) => ({
              label: (
                <Space>
                  {workspace.settings.logo_url && (
                    <img
                      src={workspace.settings.logo_url}
                      alt=""
                      style={{
                        height: '16px',
                        width: '16px',
                        marginRight: '8px',
                        objectFit: 'contain',
                        verticalAlign: 'middle'
                      }}
                    />
                  )}
                  {workspace.name}
                </Space>
              ),
              value: workspace.id
            }))}
          />
        )}
      </div>
      <div>
        <LogoutOutlined
          onClick={() => signout()}
          style={{
            fontSize: '18px',
            cursor: 'pointer',
            color: '#595959'
          }}
        />
      </div>
    </Header>
  )
}
