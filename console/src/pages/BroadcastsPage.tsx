import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Statistic,
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
import {
  broadcastApi,
  Broadcast,
  BroadcastStatus,
  BroadcastVariation
} from '../services/api/broadcast'
import { listsApi } from '../services/api/list'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faPaperPlane,
  faCirclePause,
  faCircleCheck,
  faCircleXmark,
  faPenToSquare,
  faTrashCan,
  faCirclePlay,
  faCopy,
  faEye,
  faFaceFrown
} from '@fortawesome/free-regular-svg-icons'
import {
  faArrowPointer,
  faBan,
  faChevronDown,
  faChevronUp,
  faTriangleExclamation
} from '@fortawesome/free-solid-svg-icons'
import React, { useState } from 'react'
import dayjs from '../lib/dayjs'
import { UpsertBroadcastDrawer } from '../components/broadcasts/UpsertBroadcastDrawer'
import { SendOrScheduleModal } from '../components/broadcasts/SendOrScheduleModal'
import { useAuth } from '../contexts/AuthContext'
import TemplatePreviewDrawer from '../components/templates/TemplatePreviewDrawer'

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

// Component for rendering a single A/B test variation card
interface VariationCardProps {
  variation: BroadcastVariation
  workspaceId: string
  colSpan: number
  index: number
}

const VariationCard: React.FC<VariationCardProps> = ({
  variation,
  workspaceId,
  colSpan,
  index
}) => {
  return (
    <Col span={colSpan} key={index}>
      <Card
        size="small"
        title={variation.name || variation.template?.name || `Variation ${index + 1}`}
        type="inner"
        extra={
          variation.template ? (
            <TemplatePreviewDrawer record={variation.template as any} workspaceId={workspaceId}>
              <Button size="small" type="primary" ghost>
                Preview
              </Button>
            </TemplatePreviewDrawer>
          ) : (
            <Button size="small" type="primary" ghost disabled>
              Preview
            </Button>
          )
        }
      >
        <Space direction="vertical" size="small">
          <Space>
            <Text strong>From:</Text>
            {variation.from_name || variation.template?.email?.from_name || 'N/A'}
            <span>
              &lt;
              {variation.from_email || variation.template?.email?.from_address || 'N/A'}
              &gt;
            </span>
          </Space>
          <Space>
            <Text strong>Subject:</Text>
            {variation.subject || variation.template?.email?.subject || 'N/A'}
          </Space>
          {variation.template?.email.subject_preview && (
            <Space>
              <Text strong>Subject Preview:</Text>
              {variation.template?.email?.subject_preview}
            </Space>
          )}
          {(variation.reply_to || variation.template?.email?.reply_to) && (
            <Text>Reply-to: {variation.reply_to || variation.template?.email?.reply_to}</Text>
          )}

          {variation.metrics && (
            <>
              <Divider style={{ margin: '8px 0' }} />
              <div className="grid grid-cols-3 gap-2 mt-2">
                <div>
                  <div className="font-medium text-purple-500 flex items-center">
                    <FontAwesomeIcon icon={faEye} className="mr-1" style={{ opacity: 0.7 }} /> Opens
                  </div>
                  <div>
                    {variation.metrics.opens} ({(variation.metrics.open_rate * 100).toFixed(1)}%)
                  </div>
                </div>
                <div>
                  <div className="font-medium text-cyan-500 flex items-center">
                    <FontAwesomeIcon
                      icon={faArrowPointer}
                      className="mr-1"
                      style={{ opacity: 0.7 }}
                    />{' '}
                    Clicks
                  </div>
                  <div>
                    {variation.metrics.clicks} ({(variation.metrics.click_rate * 100).toFixed(1)}%)
                  </div>
                </div>
                <div>
                  <div className="font-medium text-green-500 flex items-center">
                    <FontAwesomeIcon
                      icon={faCircleCheck}
                      className="mr-1"
                      style={{ opacity: 0.7 }}
                    />{' '}
                    Delivered
                  </div>
                  <div>
                    {variation.metrics.delivered} of {variation.metrics.recipients}
                  </div>
                </div>
              </div>
            </>
          )}
        </Space>
      </Card>
    </Col>
  )
}

// Component for rendering a single broadcast card
interface BroadcastCardProps {
  broadcast: Broadcast
  lists: any[]
  workspaceId: string
  onDelete: (broadcast: Broadcast) => void
  onPause: (broadcast: Broadcast) => void
  onResume: (broadcast: Broadcast) => void
  onCancel: (broadcast: Broadcast) => void
  onSchedule: (broadcast: Broadcast) => void
  currentWorkspace: any
  isFirst?: boolean
}

const BroadcastCard: React.FC<BroadcastCardProps> = ({
  broadcast,
  lists,
  workspaceId,
  onDelete,
  onPause,
  onResume,
  onCancel,
  onSchedule,
  currentWorkspace,
  isFirst = false
}) => {
  const [showDetails, setShowDetails] = useState(isFirst)
  console.log(broadcast)
  return (
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
                buttonContent={<FontAwesomeIcon icon={faPenToSquare} style={{ opacity: 0.7 }} />}
                buttonProps={{ size: 'small', type: 'text' }}
              />
            </Tooltip>
          )}
          {broadcast.status === 'sending' && (
            <Button type="text" size="small" onClick={() => onPause(broadcast)}>
              <Tooltip title="Pause Broadcast">
                <FontAwesomeIcon icon={faCirclePause} style={{ opacity: 0.7 }} />
              </Tooltip>
            </Button>
          )}
          {broadcast.status === 'paused' && (
            <Button type="text" size="small" onClick={() => onResume(broadcast)}>
              <Tooltip title="Resume Broadcast">
                <FontAwesomeIcon icon={faCirclePlay} style={{ opacity: 0.7 }} />
              </Tooltip>
            </Button>
          )}
          {broadcast.status === 'scheduled' && (
            <Button type="text" size="small" onClick={() => onCancel(broadcast)}>
              <Tooltip title="Cancel Broadcast">
                <FontAwesomeIcon icon={faBan} style={{ opacity: 0.7 }} />
              </Tooltip>
            </Button>
          )}
          {broadcast.status === 'draft' && (
            <>
              <Button type="text" size="small" onClick={() => onDelete(broadcast)}>
                <Tooltip title="Delete Broadcast">
                  <FontAwesomeIcon icon={faTrashCan} style={{ opacity: 0.7 }} />
                </Tooltip>
              </Button>
              <Button type="primary" size="small" ghost onClick={() => onSchedule(broadcast)}>
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
                  <FontAwesomeIcon icon={faPaperPlane} style={{ opacity: 0.7 }} /> Sent
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
                  <FontAwesomeIcon icon={faCircleCheck} style={{ opacity: 0.7 }} /> Delivered
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
                  <FontAwesomeIcon icon={faEye} style={{ opacity: 0.7 }} /> Opens
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
                  <FontAwesomeIcon
                    icon={faArrowPointer}
                    style={{ opacity: 0.7 }}
                    className="mr-1"
                  />{' '}
                  Clicks
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
                  <FontAwesomeIcon icon={faCircleXmark} style={{ opacity: 0.7 }} /> Failed
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
                  <FontAwesomeIcon icon={faTriangleExclamation} style={{ opacity: 0.7 }} /> Bounced
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
                  <FontAwesomeIcon icon={faFaceFrown} style={{ opacity: 0.7 }} /> Complaints
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
                  <FontAwesomeIcon icon={faBan} style={{ opacity: 0.7 }} /> Unsub.
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

      <div className="mt-2 text-center">
        <Button type="link" onClick={() => setShowDetails(!showDetails)}>
          {showDetails ? (
            <Space size="small">
              <FontAwesomeIcon icon={faChevronUp} style={{ opacity: 0.7 }} className="mr-1" /> Hide
              Details
            </Space>
          ) : (
            <Space size="small">
              <FontAwesomeIcon icon={faChevronDown} style={{ opacity: 0.7 }} className="mr-1" />{' '}
              Show Details
            </Space>
          )}
        </Button>
      </div>

      {showDetails && (
        <>
          <Divider />

          <Descriptions
            bordered={false}
            size="small"
            column={{ xxl: 4, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
          >
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
                <Space direction="vertical" style={{ width: '100%' }}>
                  {broadcast.audience.lists.map((listId) => {
                    const list = lists.find((l) => l.id === listId)
                    return list ? (
                      <div key={list.id}>
                        {list.name} ({list.total_active.toLocaleString()} subscribers)
                      </div>
                    ) : (
                      <div key={listId}>Unknown list ({listId})</div>
                    )
                  })}
                </Space>
              </Descriptions.Item>
            )}

            <Descriptions.Item label="Skip Duplicates">
              {broadcast.audience.skip_duplicate_emails ? (
                <FontAwesomeIcon
                  icon={faCircleCheck}
                  className="text-green-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              ) : (
                <FontAwesomeIcon
                  icon={faCircleXmark}
                  className="text-orange-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              )}
            </Descriptions.Item>

            <Descriptions.Item label="Exclude Unsubscribed">
              {broadcast.audience.exclude_unsubscribed ? (
                <FontAwesomeIcon
                  icon={faCircleCheck}
                  className="text-green-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              ) : (
                <FontAwesomeIcon
                  icon={faCircleXmark}
                  className="text-orange-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              )}
            </Descriptions.Item>

            {broadcast.audience.rate_limit_per_minute && (
              <Descriptions.Item label="Rate Limit">
                {broadcast.audience.rate_limit_per_minute}/min
              </Descriptions.Item>
            )}

            {/* Tracking Information */}
            <Descriptions.Item label="Open & Click Tracking">
              {broadcast.tracking_enabled ? (
                <FontAwesomeIcon
                  icon={faCircleCheck}
                  className="text-green-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              ) : (
                <FontAwesomeIcon
                  icon={faCircleXmark}
                  className="text-orange-500"
                  size="sm"
                  style={{ opacity: 0.7 }}
                />
              )}
            </Descriptions.Item>

            {broadcast.utm_parameters && Object.values(broadcast.utm_parameters).some((v) => v) && (
              <Descriptions.Item label="UTM Parameters">
                <Tooltip title="utm_source / utm_medium / utm_campaign">
                  <div>
                    {broadcast.utm_parameters.source &&
                      broadcast.utm_parameters.medium &&
                      broadcast.utm_parameters.campaign && (
                        <Text>
                          {broadcast.utm_parameters.source} / {broadcast.utm_parameters.medium} /{' '}
                          {broadcast.utm_parameters.campaign}
                        </Text>
                      )}
                  </div>
                </Tooltip>
              </Descriptions.Item>
            )}

            {/* Schedule Information */}
            {broadcast.schedule.is_scheduled &&
              broadcast.schedule.scheduled_date &&
              broadcast.schedule.scheduled_time && (
                <Descriptions.Item label="Scheduled">
                  {dayjs(
                    `${broadcast.schedule.scheduled_date} ${broadcast.schedule.scheduled_time}`
                  ).format('lll')}
                  {' in '}
                  {broadcast.schedule.use_recipient_timezone
                    ? 'recipients timezone'
                    : broadcast.schedule.timezone}
                </Descriptions.Item>
              )}
          </Descriptions>

          <div className="my-8">
            <Divider orientation="left">
              <Text strong>A/B Test </Text>
            </Divider>
            <Descriptions
              bordered={false}
              size="small"
              column={{ xxl: 4, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
              className="mb-4"
            >
              {!broadcast.test_settings.enabled && (
                <Descriptions.Item label="Test Sample">
                  <Badge status="warning" text="disabled" />
                </Descriptions.Item>
              )}
              {broadcast.test_settings.enabled && (
                <Descriptions.Item label="Test Sample">
                  {broadcast.test_settings.sample_percentage}%
                </Descriptions.Item>
              )}

              {broadcast.test_settings.auto_send_winner &&
                broadcast.test_settings.auto_send_winner_metric &&
                broadcast.test_settings.test_duration_hours && (
                  <Descriptions.Item label="Auto-send Winner">
                    <div className="flex items-center">
                      <FontAwesomeIcon
                        icon={faCircleCheck}
                        className="text-green-500 mr-2"
                        size="sm"
                        style={{ opacity: 0.7 }}
                      />
                      <span>
                        After {broadcast.test_settings.test_duration_hours} hours based on highest{' '}
                        {broadcast.test_settings.auto_send_winner_metric === 'open_rate'
                          ? 'opens'
                          : 'clicks'}
                      </span>
                    </div>
                  </Descriptions.Item>
                )}
            </Descriptions>

            <Row gutter={[16, 16]} className="mt-4">
              {broadcast.test_settings.variations.map((variation, index) => {
                // Calculate column width based on number of variations
                // Ensure columns are at least 6 units wide (4 per row maximum)
                const variationsCount = broadcast.test_settings.variations.length
                const colSpan = Math.max(6, Math.floor(24 / variationsCount))

                return (
                  <VariationCard
                    key={index}
                    variation={variation}
                    workspaceId={workspaceId}
                    colSpan={colSpan}
                    index={index}
                  />
                )
              })}
            </Row>
          </div>
        </>
      )}
    </Card>
  )
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
        <div className="text-2xl font-medium">Broadcasts</div>
        {currentWorkspace && hasBroadcasts && (
          <UpsertBroadcastDrawer
            workspace={currentWorkspace}
            lists={lists}
            buttonContent={<>Create Broadcast</>}
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
          {data.broadcasts.map((broadcast: Broadcast, index) => (
            <BroadcastCard
              key={broadcast.id}
              broadcast={broadcast}
              lists={lists}
              workspaceId={workspaceId}
              onDelete={openDeleteModal}
              onPause={handlePauseBroadcast}
              onResume={handleResumeBroadcast}
              onCancel={handleCancelBroadcast}
              onSchedule={handleScheduleBroadcast}
              currentWorkspace={currentWorkspace}
              isFirst={index === 0}
            />
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
                  icon={<FontAwesomeIcon icon={faCopy} style={{ opacity: 0.7 }} />}
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
