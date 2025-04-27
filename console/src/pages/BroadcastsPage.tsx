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
  Badge,
  Descriptions
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
  CopyOutlined,
  EyeOutlined,
  WarningOutlined,
  FrownOutlined,
  StopOutlined as UnsubscribeOutlined
} from '@ant-design/icons'
import { SquareMousePointer } from 'lucide-react'
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
  const [isScheduleModalVisible, setIsScheduleModalVisible] = useState(false)
  const [broadcastToSchedule, setBroadcastToSchedule] = useState<Broadcast | null>(null)
  const queryClient = useQueryClient()
  const { workspaces } = useAuth()

  // Find the current workspace from the workspaces array
  const currentWorkspace = workspaces.find((workspace) => workspace.id === workspaceId)

  const { data, isLoading } = useQuery({
    queryKey: ['broadcasts', workspaceId],
    queryFn: () => {
      return broadcastApi.list({
        workspace_id: workspaceId,
        with_templates: true
      })
    }
  })

  // Fetch lists for the current workspace
  const { data: listsData } = useQuery({
    queryKey: ['lists', workspaceId],
    queryFn: () => {
      return listsApi.list({ workspace_id: workspaceId, with_templates: true })
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
                    <Tooltip title="Edit Broadcast">
                      <UpsertBroadcastDrawer
                        workspace={currentWorkspace!}
                        broadcast={broadcast}
                        lists={lists}
                        buttonContent={<EditOutlined />}
                        buttonProps={{ size: 'small', type: 'text' }}
                      />
                    </Tooltip>
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
              <Row gutter={[16, 16]} wrap className="flex-nowrap overflow-x-auto">
                <Col span={3}>
                  <Tooltip title={`${broadcast.sent_count} total emails sent`}>
                    <Statistic
                      title={
                        <Space className="text-blue-500 font-medium">
                          <SendOutlined /> Sent
                        </Space>
                      }
                      value={broadcast.sent_count || '-'}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.delivered_count} emails successfully delivered`}>
                    <Statistic
                      title={
                        <Space className="text-green-500 font-medium">
                          <CheckCircleOutlined /> Delivered
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${((broadcast.delivered_count / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.total_opens ?? 0} total opens`}>
                    <Statistic
                      title={
                        <Space className="text-purple-500 font-medium">
                          <EyeOutlined /> Opens
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.total_opens ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.total_clicks ?? 0} total clicks`}>
                    <Statistic
                      title={
                        <Space className="text-cyan-500 font-medium">
                          <SquareMousePointer size={16} /> Clicks
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.total_clicks ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.failed_count ?? 0} emails failed to send`}>
                    <Statistic
                      title={
                        <Space className="text-orange-500 font-medium">
                          <CloseCircleOutlined /> Failed
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.failed_count ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.total_bounced ?? 0} emails bounced back`}>
                    <Statistic
                      title={
                        <Space className="text-orange-500 font-medium">
                          <WarningOutlined /> Bounced
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.total_bounced ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.total_complained ?? 0} total complaints`}>
                    <Statistic
                      title={
                        <Space className="text-orange-500 font-medium">
                          <FrownOutlined /> Complaints
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.total_complained ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
                <Col span={3}>
                  <Tooltip title={`${broadcast.total_unsubscribed ?? 0} total unsubscribes`}>
                    <Statistic
                      title={
                        <Space className="text-orange-500 font-medium">
                          <UnsubscribeOutlined /> Unsub.
                        </Space>
                      }
                      value={
                        broadcast.sent_count > 0
                          ? `${(((broadcast.total_unsubscribed ?? 0) / broadcast.sent_count) * 100).toFixed(1)}%`
                          : '-'
                      }
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Tooltip>
                </Col>
              </Row>

              <Divider />

              <Descriptions
                bordered={false}
                size="small"
                column={{ xxl: 4, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
              >
                <Descriptions.Item label="Status">
                  {getStatusBadge(broadcast.status)}
                </Descriptions.Item>

                {/* Schedule Information */}
                {broadcast.schedule.is_scheduled &&
                  broadcast.schedule.scheduled_date &&
                  broadcast.schedule.scheduled_time && (
                    <Descriptions.Item label="Scheduled">
                      {dayjs(
                        `${broadcast.schedule.scheduled_date} ${broadcast.schedule.scheduled_time}`
                      ).fromNow()}
                      {broadcast.schedule.timezone && ` (${broadcast.schedule.timezone})`}
                      {broadcast.schedule.use_recipient_timezone && (
                        <Tag className="ml-2" color="blue">
                          Uses recipient timezone
                        </Tag>
                      )}
                    </Descriptions.Item>
                  )}

                {broadcast.started_at && (
                  <Descriptions.Item label="Started">
                    {dayjs(broadcast.started_at).fromNow()}
                  </Descriptions.Item>
                )}

                {broadcast.completed_at && (
                  <Descriptions.Item label="Completed">
                    {dayjs(broadcast.completed_at).fromNow()}
                  </Descriptions.Item>
                )}

                {broadcast.paused_at && (
                  <Descriptions.Item label="Paused">
                    {dayjs(broadcast.paused_at).fromNow()}
                  </Descriptions.Item>
                )}

                {broadcast.cancelled_at && (
                  <Descriptions.Item label="Cancelled">
                    {dayjs(broadcast.cancelled_at).fromNow()}
                  </Descriptions.Item>
                )}

                {/* Audience Information */}
                {broadcast.audience.segments && broadcast.audience.segments.length > 0 && (
                  <Descriptions.Item label="Segments">
                    {broadcast.audience.segments.length} segments
                  </Descriptions.Item>
                )}

                {broadcast.audience.lists && broadcast.audience.lists.length > 0 && (
                  <Descriptions.Item label="Lists">
                    {broadcast.audience.lists.length} lists
                    {broadcast.audience.skip_duplicate_emails && (
                      <Tag className="ml-2" color="blue">
                        Skip duplicates
                      </Tag>
                    )}
                  </Descriptions.Item>
                )}

                {broadcast.audience.exclude_unsubscribed && (
                  <Descriptions.Item label="Unsubscribes">
                    <Tag color="orange">Excluded</Tag>
                  </Descriptions.Item>
                )}

                {broadcast.audience.rate_limit_per_minute && (
                  <Descriptions.Item label="Rate Limit">
                    <Tag color="purple">{broadcast.audience.rate_limit_per_minute}/min</Tag>
                  </Descriptions.Item>
                )}

                {/* Tracking Information */}
                <Descriptions.Item label="Tracking">
                  {broadcast.tracking_enabled ? <Tag color="green">On</Tag> : <Tag>Off</Tag>}
                </Descriptions.Item>

                {broadcast.utm_parameters &&
                  Object.values(broadcast.utm_parameters).some((v) => v) && (
                    <Descriptions.Item label="UTM Parameters">
                      <Tag color="cyan">Configured</Tag>
                    </Descriptions.Item>
                  )}

                {/* Test Information */}
                {broadcast.winning_variation && (
                  <Descriptions.Item label="Winning Variation">
                    <Tag color="green">{broadcast.winning_variation}</Tag>
                  </Descriptions.Item>
                )}

                {broadcast.goal_id && (
                  <Descriptions.Item label="Goal ID">
                    <Tag color="blue">{broadcast.goal_id}</Tag>
                  </Descriptions.Item>
                )}
              </Descriptions>

              {broadcast.test_settings.enabled && (
                <>
                  <Divider orientation="left">
                    <Space>
                      <CalendarOutlined className="text-purple-500" />
                      <Text strong>A/B Test Settings</Text>
                    </Space>
                  </Divider>
                  <Descriptions
                    bordered={false}
                    size="small"
                    column={{ xxl: 4, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
                    className="mb-4"
                  >
                    <Descriptions.Item label="Test Sample">
                      <Tag color="blue">{broadcast.test_settings.sample_percentage}%</Tag>
                    </Descriptions.Item>

                    {broadcast.test_settings.auto_send_winner && (
                      <Descriptions.Item label="Auto-send Winner">
                        <Tag color="green">Yes</Tag>
                      </Descriptions.Item>
                    )}

                    {broadcast.test_settings.auto_send_winner_metric && (
                      <Descriptions.Item label="Winner Metric">
                        <Tag color="purple">
                          {broadcast.test_settings.auto_send_winner_metric === 'open_rate'
                            ? 'Opens'
                            : 'Clicks'}
                        </Tag>
                      </Descriptions.Item>
                    )}

                    {broadcast.test_settings.test_duration_hours && (
                      <Descriptions.Item label="Test Duration">
                        <Tag color="cyan">{broadcast.test_settings.test_duration_hours} hours</Tag>
                      </Descriptions.Item>
                    )}

                    <Descriptions.Item label="Variations">
                      <Tag color="magenta">{broadcast.test_settings.variations.length}</Tag>
                    </Descriptions.Item>
                  </Descriptions>

                  <Row gutter={[16, 16]}>
                    {broadcast.test_settings.variations.map((variation, index) => (
                      <Col span={8} key={index}>
                        <Card
                          size="small"
                          title={variation.name || `Variation ${index + 1}`}
                          type="inner"
                        >
                          <Space direction="vertical" size="small">
                            <Text strong>Subject:</Text>
                            <Text>
                              {variation.subject || variation.template?.email?.subject || 'N/A'}
                            </Text>
                            {(variation.preview_text || variation.template?.email?.previewText) && (
                              <>
                                <Text strong>Preview:</Text>
                                <Text>
                                  {variation.preview_text || variation.template?.email?.previewText}
                                </Text>
                              </>
                            )}
                            <Text strong>From:</Text>
                            <Text>
                              {variation.from_name || variation.template?.email?.from_name || 'N/A'}
                              (
                              {variation.from_email ||
                                variation.template?.email?.from_address ||
                                'N/A'}
                              )
                            </Text>
                            {(variation.reply_to || variation.template?.email?.reply_to) && (
                              <Text>
                                Reply-to:{' '}
                                {variation.reply_to || variation.template?.email?.reply_to}
                              </Text>
                            )}
                            {variation.template && (
                              <Text type="secondary">Template: {variation.template.name}</Text>
                            )}
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
