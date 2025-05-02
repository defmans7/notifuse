import { useState } from 'react'
import {
  Card,
  Table,
  Typography,
  Spin,
  Button,
  Modal,
  Form,
  Input,
  App,
  Tag,
  Alert,
  Space
} from 'antd'
import { MailOutlined } from '@ant-design/icons'
import { WorkspaceMember } from '../services/api/types'
import { workspaceService } from '../services/api/workspace'

const { Text } = Typography

interface WorkspaceMembersProps {
  workspaceId: string
  members: WorkspaceMember[]
  loading: boolean
  onMembersChange: () => void
  isOwner: boolean
}

export function WorkspaceMembers({
  workspaceId,
  members,
  loading,
  onMembersChange,
  isOwner
}: WorkspaceMembersProps) {
  const [inviteModalVisible, setInviteModalVisible] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviting, setInviting] = useState(false)
  const { message } = App.useApp()

  // API Key Modal states
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false)
  const [apiKeyName, setApiKeyName] = useState('')
  const [creatingApiKey, setCreatingApiKey] = useState(false)
  const [apiKeyToken, setApiKeyToken] = useState('')

  const columns = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      render: (email: string) => {
        return (
          <Space>
            <Text ellipsis>{email}</Text>
          </Space>
        )
      }
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string, record: WorkspaceMember) => {
        if (record.type === 'api_key') {
          return <Tag color="purple">API Key</Tag>
        }
        return (
          <Tag color={role === 'owner' ? 'gold' : 'blue'}>
            {role.charAt(0).toUpperCase() + role.slice(1)}
          </Tag>
        )
      }
    },
    {
      title: 'Since',
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
      onMembersChange()
    } catch (error) {
      console.error('Failed to invite member', error)
      message.error('Failed to invite member')
    } finally {
      setInviting(false)
    }
  }

  const handleCreateApiKey = async () => {
    if (!apiKeyName.trim()) {
      message.error('Please enter an API key name')
      return
    }

    // Convert to snake_case
    const snakeCaseName = apiKeyName
      .trim()
      .toLowerCase()
      .replace(/\s+/g, '_')
      .replace(/[^a-z0-9_]/g, '')

    setCreatingApiKey(true)
    try {
      const response = await workspaceService.createAPIKey({
        workspace_id: workspaceId,
        email_prefix: snakeCaseName
      })

      setApiKeyToken(response.token)
      message.success('API key created successfully')

      // Refresh the members list
      onMembersChange()
    } catch (error) {
      console.error('Failed to create API key', error)
      message.error('Failed to create API key')
    } finally {
      setCreatingApiKey(false)
    }
  }

  const resetApiKeyModal = () => {
    setApiKeyModalVisible(false)
    setApiKeyName('')
    setApiKeyToken('')
  }

  const domainName = `${workspaceId}.${
    import.meta.env.VITE_API_ENDPOINT?.replace(/^https?:\/\//, '').split('/')[0] ||
    'api.example.com'
  }`

  return (
    <>
      <Card
        title="Members"
        extra={
          isOwner && (
            <Space>
              <Button type="primary" size="small" ghost onClick={() => setApiKeyModalVisible(true)}>
                Create API Key
              </Button>
              <Button type="primary" size="small" ghost onClick={() => setInviteModalVisible(true)}>
                Invite Member
              </Button>
            </Space>
          )
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
        </Form>
      </Modal>

      <Modal
        title="Create API Key"
        open={apiKeyModalVisible}
        onCancel={resetApiKeyModal}
        footer={
          apiKeyToken
            ? [
                <Button key="close" type="primary" onClick={resetApiKeyModal}>
                  Close
                </Button>
              ]
            : [
                <Button key="cancel" onClick={resetApiKeyModal}>
                  Cancel
                </Button>,
                <Button
                  key="create"
                  type="primary"
                  onClick={handleCreateApiKey}
                  loading={creatingApiKey}
                >
                  Create API Key
                </Button>
              ]
        }
      >
        {!apiKeyToken ? (
          <Form layout="vertical">
            <Form.Item
              label="API Key Name"
              required
              rules={[{ required: true, message: 'Please enter an API key name' }]}
            >
              <Input
                value={apiKeyName}
                onChange={(e) => {
                  // Convert to snake_case on change
                  const snakeCaseName = e.target.value
                    .toLowerCase()
                    .replace(/\s+/g, '_')
                    .replace(/[^a-z0-9_]/g, '')
                  setApiKeyName(snakeCaseName)
                }}
                addonAfter={'@' + domainName}
              />
            </Form.Item>
          </Form>
        ) : (
          <>
            <Alert
              message="API Key Created Successfully"
              description="This token will only be displayed once. Please save it in a secure location. It cannot be retrieved again."
              type="warning"
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Form layout="vertical">
              <Form.Item label="API Token">
                <Input.TextArea
                  value={apiKeyToken}
                  autoSize={{ minRows: 3, maxRows: 5 }}
                  readOnly
                />
              </Form.Item>
            </Form>
          </>
        )}
      </Modal>
    </>
  )
}
