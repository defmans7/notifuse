import { useState, useEffect } from 'react'
import { Card, Table, Typography, Badge, Spin, Button, Modal, Form, Input, App } from 'antd'
import { UserOutlined, PlusOutlined, MailOutlined } from '@ant-design/icons'
import { Space } from 'antd'
import { WorkspaceMember } from '../services/api/types'
import { workspaceService } from '../services/api/workspace'

const { Text } = Typography

interface WorkspaceMembersProps {
  workspaceId: string
}

export function WorkspaceMembers({ workspaceId }: WorkspaceMembersProps) {
  const [members, setMembers] = useState<WorkspaceMember[]>([])
  const [loading, setLoading] = useState(false)
  const [inviteModalVisible, setInviteModalVisible] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviting, setInviting] = useState(false)
  const { message } = App.useApp()

  useEffect(() => {
    fetchMembers()
  }, [workspaceId])

  const fetchMembers = async () => {
    setLoading(true)
    try {
      const response = await workspaceService.getMembers(workspaceId)
      setMembers(response.members)
    } catch (error) {
      console.error('Failed to fetch workspace members', error)
      message.error('Failed to fetch workspace members')
    } finally {
      setLoading(false)
    }
  }

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

  const handleInvite = async () => {
    if (!inviteEmail.trim()) {
      message.error('Please enter an email address')
      return
    }

    setInviting(true)
    try {
      // Call the API to invite the user - always with role "member"
      await workspaceService.inviteMember({
        workspace_id: workspaceId,
        email: inviteEmail,
        role: 'member' // Always set to member
      })

      message.success(`Invitation sent to ${inviteEmail}`)
      setInviteModalVisible(false)
      setInviteEmail('')

      // Refresh the members list
      fetchMembers()
    } catch (error) {
      console.error('Failed to invite member', error)
      message.error('Failed to invite member')
    } finally {
      setInviting(false)
    }
  }

  return (
    <>
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

          <p>
            <small>Invited users will join as members of the workspace.</small>
          </p>
        </Form>
      </Modal>
    </>
  )
}
