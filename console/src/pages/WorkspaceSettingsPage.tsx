import { useState, useEffect } from 'react'
import {
  Card,
  Table,
  Typography,
  Divider,
  Space,
  Badge,
  Spin,
  Form,
  Input,
  Select,
  Button,
  message,
  Modal
} from 'antd'
import { useParams } from '@tanstack/react-router'
import { workspaceService } from '../services/api/workspace'
import { WorkspaceMember, Workspace } from '../services/api/types'
import { UserOutlined, SettingOutlined, PlusOutlined, MailOutlined } from '@ant-design/icons'

const { Title, Text } = Typography
const { Option } = Select

export function WorkspaceSettingsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/settings' })
  const [members, setMembers] = useState<WorkspaceMember[]>([])
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  const [loading, setLoading] = useState(false)
  const [loadingWorkspace, setLoadingWorkspace] = useState(false)
  const [savingSettings, setSavingSettings] = useState(false)
  const [inviteModalVisible, setInviteModalVisible] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [inviting, setInviting] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    async function fetchWorkspace() {
      setLoadingWorkspace(true)
      try {
        const response = await workspaceService.get(workspaceId)
        setWorkspace(response.workspace)
        form.setFieldsValue({
          name: response.workspace.name,
          website_url: response.workspace.settings.website_url,
          timezone: response.workspace.settings.timezone
        })
      } catch (error) {
        console.error('Failed to fetch workspace', error)
      } finally {
        setLoadingWorkspace(false)
      }
    }

    fetchWorkspace()
  }, [workspaceId, form])

  useEffect(() => {
    async function fetchMembers() {
      setLoading(true)
      try {
        const response = await workspaceService.getMembers(workspaceId)
        setMembers(response.members)
      } catch (error) {
        console.error('Failed to fetch workspace members', error)
      } finally {
        setLoading(false)
      }
    }

    fetchMembers()
  }, [workspaceId])

  const columns = [
    {
      title: 'User ID',
      dataIndex: 'user_id',
      key: 'user_id',
      render: (text: string) => <Text ellipsis>{text}</Text>
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Badge
          color={role === 'owner' ? 'gold' : 'blue'}
          text={role.charAt(0).toUpperCase() + role.slice(1)}
        />
      )
    },
    {
      title: 'Member Since',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => new Date(date).toLocaleDateString()
    }
  ]

  const handleSaveSettings = async (values: any) => {
    setSavingSettings(true)
    try {
      await workspaceService.update({
        id: workspaceId,
        name: values.name,
        settings: {
          website_url: values.website_url,
          logo_url: workspace?.settings.logo_url || null,
          cover_url: workspace?.settings.cover_url || null,
          timezone: values.timezone
        }
      })

      // Refresh the workspace data
      const response = await workspaceService.get(workspaceId)
      setWorkspace(response.workspace)

      // Show success message
      message.success('Workspace settings updated successfully')
    } catch (error) {
      console.error('Failed to update workspace settings', error)
      message.error('Failed to update workspace settings')
    } finally {
      setSavingSettings(false)
    }
  }

  const handleInvite = async () => {
    if (!inviteEmail.trim()) {
      message.error('Please enter an email address')
      return
    }

    setInviting(true)
    try {
      // This is a mock implementation - you would need to implement
      // the actual API endpoint for inviting members
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success(`Invitation sent to ${inviteEmail}`)
      setInviteModalVisible(false)
      setInviteEmail('')
      setInviteRole('member')

      // Refresh members list
      const response = await workspaceService.getMembers(workspaceId)
      setMembers(response.members)
    } catch (error) {
      console.error('Failed to invite member', error)
      message.error('Failed to invite member')
    } finally {
      setInviting(false)
    }
  }

  return (
    <div className="workspace-settings">
      <Title level={2}>Workspace Settings</Title>
      <Divider />

      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <Card
          title={
            <Space>
              <SettingOutlined />
              <span>General Settings</span>
            </Space>
          }
          loading={loadingWorkspace}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={handleSaveSettings}
            initialValues={{
              name: workspace?.name || '',
              website_url: workspace?.settings.website_url || '',
              timezone: workspace?.settings.timezone || 'UTC'
            }}
          >
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
              <Button type="primary" htmlType="submit" loading={savingSettings}>
                Save Changes
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card
          title={
            <Space>
              <UserOutlined />
              <span>Workspace Members</span>
            </Space>
          }
          extra={
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setInviteModalVisible(true)}
            >
              Invite Member
            </Button>
          }
        >
          {loading ? (
            <div style={{ textAlign: 'center', padding: '20px' }}>
              <Spin />
            </div>
          ) : (
            <Table
              dataSource={members}
              columns={columns}
              rowKey="user_id"
              pagination={false}
              locale={{ emptyText: 'No members found' }}
            />
          )}
        </Card>
      </Space>

      <Modal
        title="Invite Member"
        open={inviteModalVisible}
        onCancel={() => setInviteModalVisible(false)}
        footer={[
          <Button key="cancel" onClick={() => setInviteModalVisible(false)}>
            Cancel
          </Button>,
          <Button
            key="invite"
            type="primary"
            onClick={handleInvite}
            loading={inviting}
            icon={<MailOutlined />}
          >
            Send Invitation
          </Button>
        ]}
      >
        <Form layout="vertical">
          <Form.Item
            label="Email Address"
            required
            rules={[{ required: true, message: 'Please enter an email address' }]}
          >
            <Input
              placeholder="Enter email address"
              value={inviteEmail}
              onChange={(e) => setInviteEmail(e.target.value)}
              prefix={<MailOutlined />}
            />
          </Form.Item>

          <Form.Item label="Role" required>
            <Select
              value={inviteRole}
              onChange={(value) => setInviteRole(value)}
              style={{ width: '100%' }}
            >
              <Option value="member">Member</Option>
              <Option value="owner">Owner</Option>
            </Select>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
