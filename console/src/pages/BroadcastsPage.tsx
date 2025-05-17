import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Typography,
  Space,
  Tooltip,
  Button,
  Divider,
  Modal,
  Input,
  message,
  Badge,
  Descriptions,
  Progress,
  Popover,
  Alert
} from 'antd'
import { useParams } from '@tanstack/react-router'
import {
  broadcastApi,
  Broadcast,
  BroadcastStatus,
  BroadcastVariation
} from '../services/api/broadcast'
import { listsApi } from '../services/api/list'
import { taskApi } from '../services/api/task'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faCirclePause,
  faCircleCheck,
  faCircleXmark,
  faPenToSquare,
  faTrashCan,
  faCirclePlay,
  faCopy,
  faEye,
  faCircleQuestion
} from '@fortawesome/free-regular-svg-icons'
import {
  faArrowPointer,
  faBan,
  faChevronDown,
  faChevronUp,
  faSpinner
} from '@fortawesome/free-solid-svg-icons'
import React, { useState } from 'react'
import dayjs from '../lib/dayjs'
import { UpsertBroadcastDrawer } from '../components/broadcasts/UpsertBroadcastDrawer'
import { SendOrScheduleModal } from '../components/broadcasts/SendOrScheduleModal'
import { useAuth } from '../contexts/AuthContext'
import TemplatePreviewDrawer from '../components/templates/TemplatePreviewDrawer'
import { BroadcastStats } from '../components/broadcasts/BroadcastStats'

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

  // Fetch task associated with this broadcast
  const { data: task, isLoading: isTaskLoading } = useQuery({
    queryKey: ['task', workspaceId, broadcast.id],
    queryFn: () => {
      return taskApi.findByBroadcastId(workspaceId, broadcast.id)
    },
    // Only fetch task data if the broadcast status indicates a task might exist
    enabled: ['scheduled', 'sending', 'paused', 'failed'].includes(broadcast.status),
    refetchInterval:
      broadcast.status === 'sending'
        ? 5000 // Refetch every 5 seconds for sending broadcasts
        : broadcast.status === 'scheduled'
          ? 30000 // Refetch every 30 seconds for scheduled broadcasts
          : false // Don't auto-refetch for other statuses
  })

  // Helper function to render task status badge
  const getTaskStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <Badge status="processing" text="Pending" />
      case 'running':
        return <Badge status="processing" text="Running" />
      case 'completed':
        return <Badge status="success" text="Completed" />
      case 'failed':
        return <Badge status="error" text="Failed" />
      case 'cancelled':
        return <Badge status="warning" text="Cancelled" />
      case 'paused':
        return <Badge status="warning" text="Paused" />
      default:
        return <Badge status="default" text={status} />
    }
  }

  // Create popover content for details
  const taskPopoverContent = () => {
    if (!task) return null

    return (
      <div className="max-w-xs">
        <div className="font-medium mb-2">Task Details</div>
        <div className="mb-2">
          <div className="font-medium text-gray-500">Status</div>
          <div>{getTaskStatusBadge(task.status)}</div>
        </div>

        {task.progress > 0 && (
          <div className="mb-2">
            <div className="font-medium text-gray-500">Progress</div>
            <Progress percent={Math.round(task.progress * 100)} size="small" />
          </div>
        )}

        {task.state?.message && (
          <div className="mb-2">
            <div className="font-medium text-gray-500">Message</div>
            <div>{task.state.message}</div>
          </div>
        )}

        {task.state?.send_broadcast && (
          <div className="mb-2">
            <div className="font-medium text-gray-500">Broadcast Progress</div>
            <div className="text-sm">
              Sent: {task.state.send_broadcast.sent_count} of{' '}
              {task.state.send_broadcast.total_recipients}
              {task.state.send_broadcast.failed_count > 0 && (
                <div className="text-red-500">Failed: {task.state.send_broadcast.failed_count}</div>
              )}
            </div>
          </div>
        )}

        {task.error_message && (
          <div className="mb-2">
            <div className="font-medium text-gray-500">Error</div>
            <div className="text-red-500 text-sm">{task.error_message}</div>
          </div>
        )}

        {task.type && <div className="text-xs text-gray-500 mt-2">Task type: {task.type}</div>}
      </div>
    )
  }

  return (
    <Card
      styles={{
        body: {
          padding: 0
        }
      }}
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
      key={broadcast.id}
      className="!mb-6"
    >
      <div className="p-6">
        <BroadcastStats workspaceId={workspaceId} broadcastId={broadcast.id} />
      </div>

      <div className="bg-gray-50">
        <div className="text-center py-2">
          <Button type="link" onClick={() => setShowDetails(!showDetails)}>
            {showDetails ? (
              <Space size="small">
                <FontAwesomeIcon icon={faChevronUp} style={{ opacity: 0.7 }} className="mr-1" />{' '}
                Hide Details
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
          <div className="p-6">
            {/* <Divider /> */}

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
                          {list.name} ({(list.total_active ?? 0).toLocaleString()} subscribers)
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
                    className="text-green-500 opacity-70 mt-1"
                  />
                ) : (
                  <FontAwesomeIcon
                    icon={faCircleXmark}
                    className="text-orange-500 opacity-70 mt-1"
                  />
                )}
              </Descriptions.Item>

              <Descriptions.Item label="Exclude Unsubscribed">
                {broadcast.audience.exclude_unsubscribed ? (
                  <FontAwesomeIcon
                    icon={faCircleCheck}
                    className="text-green-500 opacity-70 mt-1"
                  />
                ) : (
                  <FontAwesomeIcon
                    icon={faCircleXmark}
                    className="text-orange-500 opacity-70 mt-1"
                  />
                )}
              </Descriptions.Item>

              {broadcast.audience.rate_limit_per_minute && (
                <Descriptions.Item label="Rate Limit">
                  {broadcast.audience.rate_limit_per_minute}/min
                </Descriptions.Item>
              )}

              {broadcast.utm_parameters &&
                Object.values(broadcast.utm_parameters).some((v) => v) && (
                  <Descriptions.Item label="UTM Parameters">
                    <Tooltip title="utm_source / utm_medium / utm_campaign">
                      <div>
                        {broadcast.utm_parameters.source &&
                          broadcast.utm_parameters.medium &&
                          broadcast.utm_parameters.campaign && (
                            <Text>
                              {broadcast.utm_parameters.source} / {broadcast.utm_parameters.medium}{' '}
                              / {broadcast.utm_parameters.campaign}
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

              <Descriptions.Item label="Task Status">
                {task && (
                  <Popover
                    content={taskPopoverContent}
                    title="Task Status"
                    placement="left"
                    trigger="hover"
                  >
                    <span className="text-xs font-normal cursor-help">
                      {getTaskStatusBadge(task.status)}
                      <FontAwesomeIcon
                        icon={faCircleQuestion}
                        style={{ opacity: 0.7 }}
                        className="ml-2"
                      />
                    </span>
                  </Popover>
                )}

                {isTaskLoading && ['scheduled', 'sending', 'paused'].includes(broadcast.status) && (
                  <span className="text-xs font-normal text-gray-400">
                    <FontAwesomeIcon icon={faSpinner} spin /> Loading task...
                  </span>
                )}
              </Descriptions.Item>
            </Descriptions>

            <div className="mt-2">
              <Descriptions
                bordered={false}
                size="small"
                column={{ xxl: 4, xl: 3, lg: 2, md: 2, sm: 1, xs: 1 }}
                className="mb-4"
              >
                {!broadcast.test_settings.enabled && (
                  <Descriptions.Item label="Test Sample">
                    <Badge status="warning" text="A/B Test Disabled" />
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
          </div>
        )}
      </div>
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
  const hasMarketingEmailProvider = currentWorkspace?.settings?.marketing_email_provider_id

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

      {!hasMarketingEmailProvider && (
        <Alert
          message="Email Provider Required"
          description="You don't have a marketing email provider configured. Please set up an email provider in your workspace settings to send broadcasts."
          type="warning"
          showIcon
          className="!mb-6"
          action={
            <Button
              type="primary"
              size="small"
              href={`/workspace/${workspaceId}/settings/integrations`}
            >
              Configure Provider
            </Button>
          }
        />
      )}

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
