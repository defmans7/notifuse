import { Layout, Menu } from 'antd'
import { Outlet, Link, useParams } from '@tanstack/react-router'
import {
  MailOutlined,
  TeamOutlined,
  SettingOutlined,
  FileTextOutlined,
  LogoutOutlined
} from '@ant-design/icons'

const { Content, Sider } = Layout

export function WorkspaceLayout() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId' })

  const menuItems = [
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
      key: 'contacts',
      icon: <TeamOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/contacts" params={{ workspaceId }}>
          Contacts
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
      key: 'settings',
      icon: <SettingOutlined />,
      label: (
        <Link to="/workspace/$workspaceId/settings" params={{ workspaceId }}>
          Settings
        </Link>
      )
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: <Link to="/logout">Logout</Link>
    }
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={200} theme="light">
        <Menu
          mode="inline"
          defaultSelectedKeys={['campaigns']}
          style={{ height: '100%', borderRight: 0 }}
          items={menuItems}
        />
      </Sider>
      <Layout style={{ padding: '24px' }}>
        <Content style={{ padding: 24, margin: 0, background: '#fff' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
