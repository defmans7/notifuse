import { Layout, Menu, Select, Space, Button, Dropdown } from 'antd'
import { Outlet, Link, useParams, useMatches, useNavigate } from '@tanstack/react-router'
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faImage,
  faFolderOpen,
  faObjectGroup,
  faPaperPlane
} from '@fortawesome/free-regular-svg-icons'
import {
  faGear,
  faRightFromBracket,
  faTerminal,
  faUserGroup
} from '@fortawesome/free-solid-svg-icons'
import { useAuth } from '../contexts/AuthContext'
import { Workspace } from '../services/api/types'
import { ContactsCsvUploadProvider } from '../components/contacts/ContactsCsvUploadProvider'
import { useState } from 'react'

const { Content, Sider } = Layout

export function WorkspaceLayout() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId' })
  const { signout, workspaces, user } = useAuth()
  const navigate = useNavigate()
  const [collapsed, setCollapsed] = useState(false)

  // Use useMatches to determine the current route path
  const matches = useMatches()
  const currentPath = matches[matches.length - 1]?.pathname || ''

  // Determine which key should be selected based on the current path
  let selectedKey = 'broadcasts'
  if (currentPath.includes('/settings')) {
    selectedKey = 'settings'
  } else if (currentPath.includes('/lists')) {
    selectedKey = 'lists'
  } else if (currentPath.includes('/templates')) {
    selectedKey = 'templates'
  } else if (currentPath.includes('/contacts')) {
    selectedKey = 'contacts'
  } else if (currentPath.includes('/file-manager')) {
    selectedKey = 'file-manager'
  } else if (currentPath.includes('/transactional-notifications')) {
    selectedKey = 'transactional-notifications'
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
      icon: <FontAwesomeIcon icon={faUserGroup} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/contacts" params={{ workspaceId }}>
          Contacts
        </Link>
      )
    },
    {
      key: 'lists',
      icon: <FontAwesomeIcon icon={faFolderOpen} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/lists" params={{ workspaceId }}>
          Lists
        </Link>
      )
    },
    {
      key: 'templates',
      icon: <FontAwesomeIcon icon={faObjectGroup} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/templates" params={{ workspaceId }}>
          Templates
        </Link>
      )
    },
    {
      key: 'broadcasts',
      icon: <FontAwesomeIcon icon={faPaperPlane} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/broadcasts" params={{ workspaceId }}>
          Broadcasts
        </Link>
      )
    },
    {
      key: 'transactional-notifications',
      icon: <FontAwesomeIcon icon={faTerminal} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/transactional-notifications" params={{ workspaceId }}>
          Transactional
        </Link>
      )
    },
    {
      key: 'file-manager',
      icon: <FontAwesomeIcon icon={faImage} size="sm" style={{ opacity: 0.7 }} />,
      label: (
        <Link to="/workspace/$workspaceId/file-manager" params={{ workspaceId }}>
          File Manager
        </Link>
      )
    },
    {
      key: 'settings',
      icon: <FontAwesomeIcon icon={faGear} size="sm" style={{ opacity: 0.7 }} />,
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
          <Sider
            width={224}
            theme="light"
            style={{ position: 'relative' }}
            collapsible
            collapsed={collapsed}
            trigger={null}
          >
            <div
              style={{ padding: '24px 12px', display: 'flex', alignItems: 'center', gap: '4px' }}
            >
              {!collapsed && (
                <Select
                  value={workspaceId}
                  onChange={handleWorkspaceChange}
                  style={{ width: '100%' }}
                  placeholder="Select workspace"
                  options={workspaces.map((workspace: Workspace) => ({
                    label: (
                      <Space size="small">
                        {workspace.settings.logo_url && (
                          <img
                            src={workspace.settings.logo_url}
                            alt=""
                            style={{
                              height: '16px',
                              width: '16px',
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
              <Button
                type="text"
                icon={collapsed ? <PanelLeftOpen size={16} /> : <PanelLeftClose size={16} />}
                onClick={() => setCollapsed(!collapsed)}
                style={{ fontSize: '16px' }}
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
                position: 'fixed',
                bottom: 0,
                left: 0,
                width: collapsed ? '80px' : '224px',
                padding: '16px',
                borderTop: '1px solid #f0f0f0',
                background: '#fff',
                zIndex: 1,
                transition: 'width 0.2s'
              }}
            >
              {!collapsed && (
                <Dropdown
                  menu={{
                    items: [
                      {
                        key: 'logout',
                        label: (
                          <Space>
                            <FontAwesomeIcon
                              icon={faRightFromBracket}
                              size="sm"
                              style={{ opacity: 0.7 }}
                            />
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
              )}
              {collapsed && (
                <Button
                  type="text"
                  icon={
                    <FontAwesomeIcon icon={faRightFromBracket} size="sm" style={{ opacity: 0.7 }} />
                  }
                  onClick={() => signout()}
                  style={{ width: '100%' }}
                />
              )}
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
