import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Badge,
  Typography,
  Space,
  Tooltip,
  Button,
  Divider,
  Modal,
  Input,
  message,
  Table,
  Tag,
  Popconfirm
} from 'antd'
import { useParams } from '@tanstack/react-router'
import {
  transactionalNotificationsApi,
  TransactionalNotification,
  TransactionalStatus,
  ChannelTemplates
} from '../services/api/transactional_notifications'
import { templatesApi } from '../services/api/template'
import {
  EditOutlined,
  DeleteOutlined,
  MailOutlined,
  PlusOutlined,
  EyeOutlined,
  CopyOutlined,
  SendOutlined,
  CheckCircleOutlined,
  StopOutlined,
  TagOutlined
} from '@ant-design/icons'
import React, { useState } from 'react'
import dayjs from '../lib/dayjs'
import TemplatePreviewDrawer from '../components/templates/TemplatePreviewDrawer'
import { useAuth } from '../contexts/AuthContext'

const { Title, Paragraph, Text } = Typography

// Helper function to get status badge
const getStatusBadge = (status: TransactionalStatus) => {
  switch (status) {
    case 'draft':
      return <Badge status="default" text="Draft" />
    case 'active':
      return <Badge status="success" text="Active" />
    case 'inactive':
      return <Badge status="error" text="Inactive" />
    default:
      return <Badge status="default" text={status} />
  }
}

// Component for rendering channels
const ChannelsList: React.FC<{ channels: ChannelTemplates }> = ({ channels }) => {
  return (
    <Space direction="vertical" size="small">
      {channels.email && (
        <Tag color="blue">
          <MailOutlined /> Email
        </Tag>
      )}
      {/* Add more channel types here as they become available */}
    </Space>
  )
}

// Temporary mock components until the real ones are created
const UpsertTransactionalNotificationDrawer: React.FC<any> = ({ buttonContent, buttonProps }) => (
  <Button {...buttonProps}>{buttonContent}</Button>
)

const TestTransactionalNotificationDrawer: React.FC<any> = () => <div></div>

// Test notification modal component
const TestNotificationModal = ({
  isOpen,
  onClose,
  onSend,
  loading
}: {
  isOpen: boolean
  onClose: () => void
  onSend: (email: string) => void
  loading: boolean
}) => {
  const [email, setEmail] = useState('')

  return (
    <Modal
      title="Send Test Notification"
      open={isOpen}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="send"
          type="primary"
          onClick={() => onSend(email)}
          disabled={!email || loading}
          loading={loading}
        >
          Send Test Notification
        </Button>
      ]}
    >
      <div className="py-2">
        <p className="mb-4">Send a test notification to verify how it will look.</p>
        <Input
          placeholder="recipient@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          type="email"
        />
      </div>
    </Modal>
  )
}

export function TransactionalNotificationsPage() {
  const { workspaceId } = useParams({ strict: false })
  const { workspaces } = useAuth()
  const queryClient = useQueryClient()

  // Find the current workspace from the workspaces array
  const currentWorkspace = workspaces.find((workspace) => workspace.id === workspaceId)

  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [notificationToDelete, setNotificationToDelete] =
    useState<TransactionalNotification | null>(null)
  const [testModalOpen, setTestModalOpen] = useState(false)
  const [testLoading, setTestLoading] = useState(false)
  const [notificationToTest, setNotificationToTest] = useState<TransactionalNotification | null>(
    null
  )

  // Fetch notifications
  const {
    data: notificationsData,
    isLoading: isLoadingNotifications,
    error: notificationsError
  } = useQuery({
    queryKey: ['transactional-notifications', workspaceId],
    queryFn: () =>
      transactionalNotificationsApi.list({
        workspace_id: workspaceId as string
      }),
    enabled: !!workspaceId
  })

  // Fetch templates
  const {
    data: templatesData,
    isLoading: isLoadingTemplates,
    error: templatesError
  } = useQuery({
    queryKey: ['templates', workspaceId],
    queryFn: () =>
      templatesApi.list({
        workspace_id: workspaceId as string,
        category: 'transactional'
      }),
    enabled: !!workspaceId
  })

  const handleDeleteNotification = async () => {
    if (!notificationToDelete) return

    try {
      await transactionalNotificationsApi.delete({
        workspace_id: workspaceId as string,
        id: notificationToDelete.id
      })

      message.success('Transactional notification deleted successfully')
      setDeleteModalVisible(false)
      setNotificationToDelete(null)

      // Refresh the list
      queryClient.invalidateQueries({ queryKey: ['transactional-notifications', workspaceId] })
    } catch (error) {
      console.error('Failed to delete notification:', error)
      message.error('Failed to delete notification')
    }
  }

  const handleTestNotification = (notification: TransactionalNotification) => {
    setNotificationToTest(notification)
    setTestModalOpen(true)
  }

  const sendTestNotification = async (email: string) => {
    if (!notificationToTest) return

    setTestLoading(true)
    try {
      // Mock implementation - would need to be connected to actual API
      // await transactionalNotificationsApi.send({
      //   workspace_id: workspaceId as string,
      //   notification: {
      //     id: notificationToTest.id,
      //     contact: { email },
      //     data: { test: true }
      //   }
      // })

      // Simulate API call for now
      await new Promise((resolve) => setTimeout(resolve, 1000))

      message.success('Test notification sent successfully')
      setTestModalOpen(false)
    } catch (error) {
      message.error(`Error: ${error instanceof Error ? error.message : 'Something went wrong'}`)
    } finally {
      setTestLoading(false)
    }
  }

  if (notificationsError || templatesError) {
    return (
      <div>
        <Title level={4}>Error loading data</Title>
        <Text type="danger">
          {(notificationsError as Error)?.message || (templatesError as Error)?.message}
        </Text>
      </div>
    )
  }

  const notifications = notificationsData?.notifications || []
  const templates = templatesData?.templates || []
  const hasNotifications = notifications.length > 0

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: TransactionalNotification) => (
        <Tooltip title={'ID for API: ' + record.id}>
          <Text strong>{text}</Text>
        </Tooltip>
      )
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      render: (text: string) => <Text ellipsis>{text}</Text>
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: TransactionalStatus) => getStatusBadge(status)
    },
    {
      title: 'Channels',
      dataIndex: 'channels',
      key: 'channels',
      render: (channels: ChannelTemplates) => <ChannelsList channels={channels} />
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => (
        <Tooltip
          title={
            dayjs(date).tz(currentWorkspace?.settings.timezone).format('llll') +
            ' in ' +
            currentWorkspace?.settings.timezone
          }
        >
          <span>{dayjs(date).format('ll')}</span>
        </Tooltip>
      )
    },
    {
      title: 'Public',
      dataIndex: 'is_public',
      key: 'is_public',
      render: (isPublic: boolean) => (isPublic ? 'Yes' : 'No')
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: TransactionalNotification) => (
        <Space>
          <Tooltip title="Edit">
            <UpsertTransactionalNotificationDrawer
              workspace={currentWorkspace}
              notification={record}
              templates={templates}
              buttonContent={<EditOutlined />}
              buttonProps={{ type: 'text', size: 'small' }}
            />
          </Tooltip>
          <Tooltip title="Test">
            <Button type="text" size="small" onClick={() => handleTestNotification(record)}>
              <SendOutlined />
            </Button>
          </Tooltip>
          <Tooltip title="Delete">
            <Popconfirm
              title="Delete the notification?"
              description="Are you sure you want to delete this notification? This cannot be undone."
              onConfirm={() => handleDeleteNotification()}
              okText="Yes, Delete"
              cancelText="Cancel"
              placement="topRight"
            >
              <Button type="text" size="small">
                <DeleteOutlined />
              </Button>
            </Popconfirm>
          </Tooltip>
        </Space>
      )
    }
  ]

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <div className="text-2xl font-medium">Transactional Notifications</div>
        {currentWorkspace && hasNotifications && (
          <UpsertTransactionalNotificationDrawer
            workspace={currentWorkspace}
            templates={templates}
            buttonContent={
              <Space>
                <PlusOutlined /> New Notification
              </Space>
            }
            buttonProps={{ type: 'primary' }}
          />
        )}
      </div>

      {isLoadingNotifications || isLoadingTemplates ? (
        <Table columns={columns} dataSource={[]} loading={true} rowKey="id" />
      ) : hasNotifications ? (
        <Table
          columns={columns}
          dataSource={notifications}
          rowKey="id"
          pagination={{ hideOnSinglePage: true }}
        />
      ) : (
        <div className="text-center py-12">
          <Title level={4} type="secondary">
            No transactional notifications found
          </Title>
          <Paragraph type="secondary">Create your first notification to get started</Paragraph>
          <div className="mt-4">
            {currentWorkspace && (
              <UpsertTransactionalNotificationDrawer
                workspace={currentWorkspace}
                templates={templates}
                buttonContent="Create Notification"
                buttonProps={{ size: 'large' }}
              />
            )}
          </div>
        </div>
      )}

      {/* Test notification modal */}
      <TestNotificationModal
        isOpen={testModalOpen}
        onClose={() => setTestModalOpen(false)}
        onSend={sendTestNotification}
        loading={testLoading}
      />
    </div>
  )
}
