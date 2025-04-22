import { Layout, Menu, Select, Space, Button, Dropdown } from 'antd'
import { Outlet, Link, useParams, useMatches, useNavigate } from '@tanstack/react-router'
import {
  MailOutlined,
  TeamOutlined,
  SettingOutlined,
  FileTextOutlined,
  LogoutOutlined,
  FolderOpenOutlined,
  PictureOutlined
} from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'
import { Workspace } from '../services/api/types'
import logo from '../assets/logo.png'
import { ContactsCsvUploadProvider } from '../components/contacts/ContactsCsvUploadProvider'

const { Content, Sider } = Layout

export function WorkspaceLayout() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId' })
  const { signout, workspaces, user } = useAuth()
  const navigate = useNavigate()

  // Use useMatches to determine the current route path
  const matches = useMatches()
  const currentPath = matches[matches.length - 1]?.pathname || ''

  // Determine which key should be selected based on the current path
  let selectedKey = 'campaigns'
  if (currentPath.includes('/settings')) {
    selectedKey = 'settings'
  } else if (currentPath.includes('/lists')) {
    selectedKey = 'lists'
  } else if (currentPath.includes('/templates')) {
    selectedKey = 'templates'
  } else if (currentPath.includes('/contacts')) {
    selectedKey = 'contacts'
  } else if (currentPath.includes('/media')) {
    selectedKey = 'media'
  }

  const handleWorkspaceChange = (workspaceId: string) => {
    navigate({
      to: '/workspace/$workspaceId/contacts',
      params: { workspaceId }
    })
  }

  const menuItems = [
    {
      key: 'contacts',
      icon: <TeamOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/contacts" params={{ workspaceId }}>
          Contacts
        </Link>
      )
    },
    {
      key: 'lists',
      icon: <FolderOpenOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/lists" params={{ workspaceId }}>
          Lists
        </Link>
      )
    },
    {
      key: 'templates',
      icon: <FileTextOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/templates" params={{ workspaceId }}>
          Templates
        </Link>
      )
    },
    {
      key: 'campaigns',
      icon: <MailOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/campaigns" params={{ workspaceId }}>
          Campaigns
        </Link>
      )
    },
    {
      key: 'media',
      icon: <PictureOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/media" params={{ workspaceId }}>
          Media
        </Link>
      )
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/settings" params={{ workspaceId }}>
          Settings
        </Link>
      )
    }
  ]

  return (
    <ContactsCsvUploadProvider workspaceId={workspaceId}>
      <Layout style={{ minHeight: '100vh' }}>
        <Layout>
          <Sider width={250} theme="light" style={{ position: 'relative' }}>
            <div style={{ padding: '16px 0 0 24px' }}>
              <img
                src={logo}
                alt="Notifuse"
                style={{ height: '32px', cursor: 'pointer' }}
                onClick={() => navigate({ to: '/' })}
              />
            </div>
            <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
              <Select
                value={workspaceId}
                onChange={handleWorkspaceChange}
                style={{ width: '100%' }}
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
            </div>
            <Menu
              mode="inline"
              selectedKeys={[selectedKey]}
              style={{ height: 'calc(100% - 120px)', borderRight: 0 }}
              items={menuItems}
            />
            <div
              style={{
                position: 'absolute',
                bottom: 0,
                left: 0,
                right: 0,
                padding: '16px',
                borderTop: '1px solid #f0f0f0',
                background: '#fff'
              }}
            >
              <Dropdown
                menu={{
                  items: [
                    {
                      key: 'logout',
                      label: (
                        <Space>
                          <LogoutOutlined />
                          Logout
                        </Space>
                      ),
                      onClick: () => signout()
                    }
                  ]
                }}
                trigger={['click']}
                placement="bottomRight"
              >
                <Button type="text" block>
                  <div style={{ padding: '4px 8px', color: '#595959', cursor: 'pointer' }}>
                    {user?.email}
                  </div>
                </Button>
              </Dropdown>
            </div>
          </Sider>
          <Layout style={{ padding: '24px' }}>
            <Content>
              <Outlet />
            </Content>
          </Layout>
        </Layout>
      </Layout>
    </ContactsCsvUploadProvider>
  )
}
