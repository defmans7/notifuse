import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Typography,
  Space,
  Tooltip,
  Button,
  Divider,
  Modal,
  Input,
  message,
  Badge
} from 'antd'
import { useParams } from '@tanstack/react-router'
import { broadcastApi, Broadcast, BroadcastStatus } from '../services/api/broadcast'
import { listsApi } from '../services/api/list'
import {
  CalendarOutlined,
  SendOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  MailOutlined,
  PlayCircleOutlined,
  StopOutlined,
  CopyOutlined
} from '@ant-design/icons'
import { useState } from 'react'
import dayjs from '../lib/dayjs'
import { UpsertBroadcastDrawer } from '../components/broadcasts/UpsertBroadcastDrawer'
import { SendOrScheduleModal } from '../components/broadcasts/SendOrScheduleModal'
import { useAuth } from '../contexts/AuthContext'

const { Title, Paragraph, Text } = Typography

// Helper function to get status badge
const getStatusBadge = (status: BroadcastStatus) => {
  switch (status) {
    case 'draft':
      return <Badge status="default" text="Draft" />
    case 'scheduled':
      return <Badge status="processing" text="Scheduled" />
    case 'sending':
      return <Badge status="processing" text="Sending" />
    case 'paused':
      return <Badge status="warning" text="Paused" />
    case 'sent':
      return <Badge status="success" text="Sent" />
    case 'cancelled':
      return <Badge status="error" text="Cancelled" />
    case 'failed':
      return <Badge status="error" text="Failed" />
    default:
      return <Badge status="default" text={status} />
  }
}

export function BroadcastsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/broadcasts' })
  const [deleteModalVisible, setDeleteModalVisible] = useState(false)
  const [broadcastToDelete, setBroadcastToDelete] = useState<Broadcast | null>(null)
  const [confirmationInput, setConfirmationInput] = useState('')
  const [isDeleting, setIsDeleting] = useState(false)
  const [selectedBroadcast, setSelectedBroadcast] = useState<Broadcast | null>(null)
  const [isScheduleModalVisible, setIsScheduleModalVisible] = useState(false)
  const [broadcastToSchedule, setBroadcastToSchedule] = useState<Broadcast | null>(null)
  const queryClient = useQueryClient()
  const { workspaces } = useAuth()

  // Find the current workspace from the workspaces array
  const currentWorkspace = workspaces.find((workspace) => workspace.id === workspaceId)

  const { data, isLoading } = useQuery({
    queryKey: ['broadcasts', workspaceId],
    queryFn: () => {
      return broadcastApi.list({ workspace_id: workspaceId })
    }
  })

  // Fetch lists for the current workspace
  const { data: listsData } = useQuery({
    queryKey: ['lists', workspaceId],
    queryFn: () => {
      return listsApi.list({ workspace_id: workspaceId })
    }
  })

  const lists = listsData?.lists || []

  const handleDeleteBroadcast = async () => {
    if (!broadcastToDelete) return

    setIsDeleting(true)
    try {
      await broadcastApi.delete({
        workspace_id: workspaceId,
        id: broadcastToDelete.id
      })

      message.success(`Broadcast "${broadcastToDelete.name}" deleted successfully`)
      queryClient.invalidateQueries({ queryKey: ['broadcasts', workspaceId] })
      setDeleteModalVisible(false)
      setBroadcastToDelete(null)
      setConfirmationInput('')
    } catch (error) {
      message.error('Failed to delete broadcast')
      console.error(error)
    } finally {
      setIsDeleting(false)
    }
  }

  const handlePauseBroadcast = async (broadcast: Broadcast) => {
    try {
      await broadcastApi.pause({
        workspace_id: workspaceId,
        id: broadcast.id
      })
      message.success(`Broadcast "${broadcast.name}" paused successfully`)
      queryClient.invalidateQueries({ queryKey: ['broadcasts', workspaceId] })
    } catch (error) {
      message.error('Failed to pause broadcast')
      console.error(error)
    }
  }

  const handleResumeBroadcast = async (broadcast: Broadcast) => {
    try {
      await broadcastApi.resume({
        workspace_id: workspaceId,
        id: broadcast.id
      })
      message.success(`Broadcast "${broadcast.name}" resumed successfully`)
      queryClient.invalidateQueries({ queryKey: ['broadcasts', workspaceId] })
    } catch (error) {
      message.error('Failed to resume broadcast')
      console.error(error)
    }
  }

  const handleCancelBroadcast = async (broadcast: Broadcast) => {
    try {
      await broadcastApi.cancel({
        workspace_id: workspaceId,
        id: broadcast.id
      })
      message.success(`Broadcast "${broadcast.name}" cancelled successfully`)
      queryClient.invalidateQueries({ queryKey: ['broadcasts', workspaceId] })
    } catch (error) {
      message.error('Failed to cancel broadcast')
      console.error(error)
    }
  }

  const openDeleteModal = (broadcast: Broadcast) => {
    setBroadcastToDelete(broadcast)
    setDeleteModalVisible(true)
  }

  const closeDeleteModal = () => {
    setDeleteModalVisible(false)
    setBroadcastToDelete(null)
    setConfirmationInput('')
  }

  const handleEditBroadcast = (broadcast: Broadcast) => {
    setSelectedBroadcast(broadcast)
  }

  const handleCloseDrawer = () => {
    setSelectedBroadcast(null)
  }

  const handleScheduleBroadcast = (broadcast: Broadcast) => {
    setBroadcastToSchedule(broadcast)
    setIsScheduleModalVisible(true)
  }

  const closeScheduleModal = () => {
    setIsScheduleModalVisible(false)
    setBroadcastToSchedule(null)
  }

  const hasBroadcasts = !isLoading && data?.broadcasts && data.broadcasts.length > 0

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2}>Broadcasts</Title>
        {currentWorkspace && hasBroadcasts && (
          <UpsertBroadcastDrawer
            workspace={currentWorkspace}
            broadcast={selectedBroadcast || undefined}
            onClose={handleCloseDrawer}
            lists={lists}
            buttonContent={
              <Space>
                <MailOutlined />
                Create Broadcast
              </Space>
            }
          />
        )}
      </div>

      {isLoading ? (
        <Row gutter={[16, 16]}>
          {[1, 2, 3].map((key) => (
            <Col xs={24} sm={12} lg={8} key={key}>
              <Card loading variant="outlined" />
            </Col>
          ))}
        </Row>
      ) : hasBroadcasts ? (
        <div>
          {data.broadcasts.map((broadcast: Broadcast) => (
            <Card
              title={
                <Space size="large">
                  <div>{broadcast.name}</div>
                  <div className="text-xs font-normal">{getStatusBadge(broadcast.status)}</div>
                </Space>
              }
              extra={
                <Space>
                  {(broadcast.status === 'draft' || broadcast.status === 'scheduled') && (
                    <Button type="text" size="small" onClick={() => handleEditBroadcast(broadcast)}>
                      <Tooltip title="Edit Broadcast">
                        <EditOutlined />
                      </Tooltip>
                    </Button>
                  )}
                  {broadcast.status === 'sending' && (
                    <Button
                      type="text"
                      size="small"
                      onClick={() => handlePauseBroadcast(broadcast)}
                    >
                      <Tooltip title="Pause Broadcast">
                        <PauseCircleOutlined />
                      </Tooltip>
                    </Button>
                  )}
                  {broadcast.status === 'paused' && (
                    <Button
                      type="text"
                      size="small"
                      onClick={() => handleResumeBroadcast(broadcast)}
                    >
                      <Tooltip title="Resume Broadcast">
                        <PlayCircleOutlined />
                      </Tooltip>
                    </Button>
                  )}
                  {broadcast.status === 'scheduled' && (
                    <Button
                      type="text"
                      size="small"
                      onClick={() => handleCancelBroadcast(broadcast)}
                    >
                      <Tooltip title="Cancel Broadcast">
                        <StopOutlined />
                      </Tooltip>
                    </Button>
                  )}
                  {broadcast.status === 'draft' && (
                    <>
                      <Button type="text" size="small" onClick={() => openDeleteModal(broadcast)}>
                        <Tooltip title="Delete Broadcast">
                          <DeleteOutlined />
                        </Tooltip>
                      </Button>
                      <Button
                        type="primary"
                        size="small"
                        ghost
                        onClick={() => handleScheduleBroadcast(broadcast)}
                      >
                        Send or Schedule
                      </Button>
                    </>
                  )}
                </Space>
              }
              variant="borderless"
              key={broadcast.id}
              className="!mb-6"
            >
              <Row gutter={[16, 16]} wrap={false}>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <SendOutlined className="text-blue-500" /> Sent
                      </Space>
                    }
                    value={broadcast.sent_count}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <CheckCircleOutlined className="text-green-500" /> Delivered
                      </Space>
                    }
                    value={broadcast.delivered_count}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <CloseCircleOutlined className="text-red-500" /> Failed
                      </Space>
                    }
                    value={broadcast.failed_count}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
              </Row>

              <Divider />

              <Row gutter={16}>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    <Text type="secondary">Status: {getStatusBadge(broadcast.status)}</Text>
                    {broadcast.scheduled_at && (
                      <Text type="secondary">
                        Scheduled: {dayjs(broadcast.scheduled_at).fromNow()}
                      </Text>
                    )}
                    {broadcast.started_at && (
                      <Text type="secondary">Started: {dayjs(broadcast.started_at).fromNow()}</Text>
                    )}
                    {broadcast.completed_at && (
                      <Text type="secondary">
                        Completed: {dayjs(broadcast.completed_at).fromNow()}
                      </Text>
                    )}
                  </Space>
                </Col>
                <Col span={12}>
                  <Space direction="vertical" size="small">
                    {broadcast.audience.segments && broadcast.audience.segments.length > 0 && (
                      <Text>Audience: {broadcast.audience.segments.length} segments</Text>
                    )}
                    {broadcast.audience.lists && broadcast.audience.lists.length > 0 && (
                      <Text>Audience: {broadcast.audience.lists.length} lists</Text>
                    )}
                    <Text>
                      Tracking:{' '}
                      {broadcast.tracking_enabled ? <Tag color="green">On</Tag> : <Tag>Off</Tag>}
                    </Text>
                    {broadcast.winning_variation && (
                      <Text>
                        Winning Variation: <Tag color="green">{broadcast.winning_variation}</Tag>
                      </Text>
                    )}
                  </Space>
                </Col>
              </Row>

              {broadcast.test_settings.enabled && (
                <>
                  <Divider orientation="left">
                    <Space>
                      <CalendarOutlined className="text-purple-500" />
                      <Text strong>A/B Test Variations</Text>
                    </Space>
                  </Divider>
                  <Row gutter={[16, 16]}>
                    {broadcast.test_settings.variations.map((variation, index) => (
                      <Col span={8} key={index}>
                        <Card size="small" title={`Variation ${index + 1}`} bordered={false}>
                          <Space direction="vertical" size="small">
                            <Text>{`Template ${index + 1}`}</Text>
                            <Button size="small" type="primary" ghost>
                              Preview
                            </Button>
                          </Space>
                        </Card>
                      </Col>
                    ))}
                  </Row>
                </>
              )}
            </Card>
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <Title level={4} type="secondary">
            No broadcasts found
          </Title>
          <Paragraph type="secondary">Create your first broadcast to get started</Paragraph>
          <div className="mt-4">
            {currentWorkspace && (
              <UpsertBroadcastDrawer
                workspace={currentWorkspace}
                buttonProps={{ size: 'large', icon: <MailOutlined /> }}
                lists={lists}
                buttonContent="Create Broadcast"
              />
            )}
          </div>
        </div>
      )}

      <SendOrScheduleModal
        broadcast={broadcastToSchedule}
        visible={isScheduleModalVisible}
        onClose={closeScheduleModal}
        workspaceId={workspaceId}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: ['broadcasts', workspaceId] })
        }}
      />

      <Modal
        title="Delete Broadcast"
        open={deleteModalVisible}
        onCancel={closeDeleteModal}
        footer={[
          <Button key="cancel" onClick={closeDeleteModal}>
            Cancel
          </Button>,
          <Button
            key="delete"
            type="primary"
            danger
            loading={isDeleting}
            disabled={confirmationInput !== (broadcastToDelete?.id || '')}
            onClick={handleDeleteBroadcast}
          >
            Delete
          </Button>
        ]}
      >
        {broadcastToDelete && (
          <>
            <p>Are you sure you want to delete the broadcast "{broadcastToDelete.name}"?</p>
            <p>
              This action cannot be undone. To confirm, please enter the broadcast ID:{' '}
              <Text code>{broadcastToDelete.id}</Text>
              <Tooltip title="Copy to clipboard">
                <Button
                  type="text"
                  icon={<CopyOutlined />}
                  size="small"
                  onClick={() => {
                    navigator.clipboard.writeText(broadcastToDelete.id)
                    message.success('Broadcast ID copied to clipboard')
                  }}
                />
              </Tooltip>
            </p>
            <Input
              placeholder="Enter broadcast ID to confirm"
              value={confirmationInput}
              onChange={(e) => setConfirmationInput(e.target.value)}
              status={
                confirmationInput && confirmationInput !== broadcastToDelete.id ? 'error' : ''
              }
            />
            {confirmationInput && confirmationInput !== broadcastToDelete.id && (
              <p className="text-red-500 mt-2">ID doesn't match</p>
            )}
          </>
        )}
      </Modal>
    </div>
  )
}
