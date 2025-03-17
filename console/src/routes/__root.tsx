import { Layout, Button, Menu } from 'antd'
import { Outlet, useNavigate, Link, useMatchRoute } from '@tanstack/react-router'
import {
  LogoutOutlined,
  DashboardOutlined,
  TeamOutlined,
  FileTextOutlined,
  SettingOutlined,
  SendOutlined
} from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'
import { createRootRoute } from '@tanstack/react-router'

const { Header, Content, Sider } = Layout

function Root() {
  const { user, workspaces } = useAuth()
  const navigate = useNavigate()
  const matchRoute = useMatchRoute()
  const publicRoutes = ['/signin', '/accept-invitation']
  const isPublicRoute = publicRoutes.includes(window.location.pathname)

  const isInWorkspace =
    matchRoute({ to: '/workspace/$workspaceId' }) ||
    matchRoute({ to: '/workspace/$workspaceId/templates' }) ||
    matchRoute({ to: '/workspace/$workspaceId/settings' }) ||
    matchRoute({ to: '/workspace/$workspaceId/contacts' }) ||
    matchRoute({ to: '/workspace/$workspaceId/campaigns' })

  // Get the workspace ID from the URL
  const workspaceId = window.location.pathname.split('/workspace/')[1]?.split('/')[0]
  const workspace = workspaceId ? workspaces.find((w) => w.id === workspaceId) : null

  const handleLogout = () => {
    navigate({ to: '/logout' })
  }

  if (isPublicRoute) {
    return <Outlet />
  }

  if (isInWorkspace && !workspace) {
    navigate({ to: '/' })
    return null
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Link to="/" style={{ color: 'white', fontSize: '18px', marginRight: '24px' }}>
            Notifuse
          </Link>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
          {user && <span style={{ color: 'white' }}>{user.email}</span>}
          <Button icon={<LogoutOutlined />} onClick={handleLogout}>
            Logout
          </Button>
        </div>
      </Header>
      <Layout>
        {isInWorkspace && workspace && (
          <Sider width={200}>
            <Menu
              mode="inline"
              defaultSelectedKeys={['1']}
              style={{ height: '100%', borderRight: 0 }}
              items={[
                {
                  key: '1',
                  icon: <DashboardOutlined />,
                  label: (
                    <Link to="/workspace/$workspaceId" params={{ workspaceId: workspace.id }}>
                      Dashboard
                    </Link>
                  )
                },
                {
                  key: '2',
                  icon: <TeamOutlined />,
                  label: (
                    <Link
                      to="/workspace/$workspaceId/contacts"
                      params={{ workspaceId: workspace.id }}
                    >
                      Contacts
                    </Link>
                  )
                },
                {
                  key: '3',
                  icon: <FileTextOutlined />,
                  label: (
                    <Link
                      to="/workspace/$workspaceId/templates"
                      params={{ workspaceId: workspace.id }}
                    >
                      Templates
                    </Link>
                  )
                },
                {
                  key: '4',
                  icon: <SendOutlined />,
                  label: (
                    <Link
                      to="/workspace/$workspaceId/campaigns"
                      params={{ workspaceId: workspace.id }}
                    >
                      Campaigns
                    </Link>
                  )
                },
                {
                  key: '5',
                  icon: <SettingOutlined />,
                  label: (
                    <Link
                      to="/workspace/$workspaceId/settings"
                      params={{ workspaceId: workspace.id }}
                    >
                      Settings
                    </Link>
                  )
                }
              ]}
            />
          </Sider>
        )}
        <Layout style={{ padding: '24px' }}>
          <Content
            style={{
              padding: 24,
              margin: 0,
              background: '#fff',
              minHeight: 'calc(100vh - 112px)' // 64px header + 24px * 2 padding
            }}
          >
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  )
}

export const Route = createRootRoute({
  component: Root
})
