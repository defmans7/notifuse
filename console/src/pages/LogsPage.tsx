import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Space,
  Tabs,
  Table,
  Tag,
  Button,
  Select,
  Input,
  Card,
  Tooltip,
  Popover
} from 'antd'
import { useParams } from '@tanstack/react-router'
import { listMessages, MessageHistory, MessageStatus } from '../services/api/messages_history'
import { useAuth } from '../contexts/AuthContext'
import dayjs from '../lib/dayjs'
import React, { useState, useMemo } from 'react'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faEnvelope,
  faCircleCheck,
  faCircleXmark,
  faEye,
  faHandPointer
} from '@fortawesome/free-regular-svg-icons'
import {
  faTriangleExclamation,
  faBan,
  faArrowRightFromBracket
} from '@fortawesome/free-solid-svg-icons'
import { getWebhookStatus } from '../services/api/webhook_registration'

const { Title, Text } = Typography

// Define status icon and color mappings
const statusConfig: Record<MessageStatus, { icon: React.ReactNode; color: string; label: string }> =
  {
    sent: {
      icon: <FontAwesomeIcon icon={faEnvelope} />,
      color: 'blue',
      label: 'Sent'
    },
    delivered: {
      icon: <FontAwesomeIcon icon={faCircleCheck} />,
      color: 'green',
      label: 'Delivered'
    },
    failed: {
      icon: <FontAwesomeIcon icon={faCircleXmark} />,
      color: 'red',
      label: 'Failed'
    },
    opened: {
      icon: <FontAwesomeIcon icon={faEye} />,
      color: 'purple',
      label: 'Opened'
    },
    clicked: {
      icon: <FontAwesomeIcon icon={faHandPointer} />,
      color: 'geekblue',
      label: 'Clicked'
    },
    bounced: {
      icon: <FontAwesomeIcon icon={faTriangleExclamation} />,
      color: 'orange',
      label: 'Bounced'
    },
    complained: {
      icon: <FontAwesomeIcon icon={faBan} />,
      color: 'volcano',
      label: 'Complained'
    },
    unsubscribed: {
      icon: <FontAwesomeIcon icon={faArrowRightFromBracket} />,
      color: 'gold',
      label: 'Unsubscribed'
    }
  }

// Simple filter field type
interface FilterOption {
  key: string
  label: string
  options?: { value: string; label: string }[]
}

// Define filter fields for message history
const filterOptions: FilterOption[] = [
  {
    key: 'status',
    label: 'Status',
    options: Object.entries(statusConfig).map(([value, { label }]) => ({
      value,
      label
    }))
  },
  {
    key: 'channel',
    label: 'Channel',
    options: [
      { value: 'email', label: 'Email' },
      { value: 'sms', label: 'SMS' },
      { value: 'push', label: 'Push' }
    ]
  },
  { key: 'contact_email', label: 'Contact Email' },
  { key: 'template_id', label: 'Template ID' },
  { key: 'broadcast_id', label: 'Broadcast ID' },
  {
    key: 'has_error',
    label: 'Has Error',
    options: [
      { value: 'true', label: 'With Errors' },
      { value: 'false', label: 'No Errors' }
    ]
  }
]

// Simple filter interface
interface Filter {
  field: string
  value: string
  label: string
}

// Messages History Tab
const MessagesHistoryTab: React.FC<{ workspaceId: string }> = ({ workspaceId }) => {
  const { workspaces } = useAuth()
  const [currentCursor, setCurrentCursor] = useState<string | undefined>(undefined)
  const [allMessages, setAllMessages] = useState<MessageHistory[]>([])
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const queryClient = useQueryClient()

  // State for filters
  const [activeFilters, setActiveFilters] = useState<Filter[]>([])
  const [openPopovers, setOpenPopovers] = useState<Record<string, boolean>>({})
  const [tempFilterValues, setTempFilterValues] = useState<Record<string, string>>({})

  // Create API filters from active filters
  const apiFilters = useMemo(() => {
    return activeFilters.reduce(
      (filters, filter) => {
        const { field, value } = filter

        // Special case for has_error which needs to be converted to boolean
        if (field === 'has_error') {
          filters[field] = value === 'true'
        } else {
          filters[field] = value
        }

        return filters
      },
      {} as Record<string, any>
    )
  }, [activeFilters])

  // Find the current workspace from the workspaces array
  const currentWorkspace = workspaces.find((workspace) => workspace.id === workspaceId)
  const timezone = currentWorkspace?.settings.timezone || 'UTC'

  // Load initial filters from URL on mount
  React.useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search)
    const initialFilters: Filter[] = []

    filterOptions.forEach((option) => {
      const value = searchParams.get(option.key)
      if (value) {
        initialFilters.push({
          field: option.key,
          value,
          label: option.label
        })
      }
    })

    if (initialFilters.length > 0) {
      setActiveFilters(initialFilters)
    }
  }, [])

  // Update URL when filters change
  React.useEffect(() => {
    const searchParams = new URLSearchParams()

    activeFilters.forEach((filter) => {
      searchParams.set(filter.field, filter.value)
    })

    const newUrl =
      window.location.pathname + (searchParams.toString() ? `?${searchParams.toString()}` : '')

    window.history.pushState({ path: newUrl }, '', newUrl)
  }, [activeFilters])

  // Fetch message history
  const {
    data: messagesData,
    isLoading,
    error
  } = useQuery({
    queryKey: ['messages-history', workspaceId, apiFilters, currentCursor],
    queryFn: async () => {
      return listMessages(workspaceId, {
        ...apiFilters,
        limit: 20,
        cursor: currentCursor
      })
    },
    staleTime: 5000,
    refetchOnWindowFocus: false
  })

  // Reset the cursor and accumulated messages when filters change
  React.useEffect(() => {
    setAllMessages([])
    setCurrentCursor(undefined)
    queryClient.resetQueries({ queryKey: ['messages-history', workspaceId] })
  }, [apiFilters, workspaceId, queryClient])

  // Update allMessages when data changes
  React.useEffect(() => {
    // If data is still loading or not available, don't update
    if (isLoading || !messagesData) return

    if (messagesData.messages) {
      if (!currentCursor) {
        // Initial load or filter change - replace all messages
        setAllMessages(messagesData.messages)
      } else if (messagesData.messages.length > 0) {
        // If we have a cursor and new messages, append them
        setAllMessages((prev) => [...prev, ...messagesData.messages])
      }
    }

    // Reset loading more flag
    setIsLoadingMore(false)
  }, [messagesData, currentCursor, isLoading])

  // Load more messages
  const handleLoadMore = () => {
    if (messagesData?.next_cursor) {
      setIsLoadingMore(true)
      setCurrentCursor(messagesData.next_cursor)
    }
  }

  // Handle applying a filter
  const applyFilter = (field: string, value: string) => {
    // Remove any existing filter for this field
    const updatedFilters = activeFilters.filter((f) => f.field !== field)

    // Add the new filter if it has a value
    if (value) {
      const filterOption = filterOptions.find((option) => option.key === field)
      if (filterOption) {
        updatedFilters.push({
          field,
          value,
          label: filterOption.label
        })
      }
    }

    setActiveFilters(updatedFilters)
    setOpenPopovers({ ...openPopovers, [field]: false })
  }

  // Handle clearing a filter
  const clearFilter = (field: string) => {
    setActiveFilters(activeFilters.filter((f) => f.field !== field))
    setTempFilterValues({ ...tempFilterValues, [field]: '' })
    setOpenPopovers({ ...openPopovers, [field]: false })
  }

  // Clear all filters
  const clearAllFilters = () => {
    setActiveFilters([])
    setTempFilterValues({})
    // Clear URL params
    window.history.pushState({ path: window.location.pathname }, '', window.location.pathname)
  }

  // Render filter buttons
  const renderFilterButtons = () => {
    return (
      <Space wrap>
        {filterOptions.map((option) => {
          const isActive = activeFilters.some((f) => f.field === option.key)
          const activeFilter = activeFilters.find((f) => f.field === option.key)

          return (
            <Popover
              key={option.key}
              trigger="click"
              open={openPopovers[option.key]}
              onOpenChange={(visible) => {
                // Initialize temp value when opening
                if (visible && activeFilter) {
                  setTempFilterValues({
                    ...tempFilterValues,
                    [option.key]: activeFilter.value
                  })
                }
                setOpenPopovers({ ...openPopovers, [option.key]: visible })
              }}
              content={
                <div style={{ width: 200 }}>
                  {option.options ? (
                    <Select
                      style={{ width: '100%', marginBottom: 8 }}
                      placeholder={`Select ${option.label}`}
                      value={tempFilterValues[option.key] || undefined}
                      onChange={(value) =>
                        setTempFilterValues({
                          ...tempFilterValues,
                          [option.key]: value
                        })
                      }
                      options={option.options}
                      allowClear
                    />
                  ) : (
                    <Input
                      placeholder={`Enter ${option.label}`}
                      value={tempFilterValues[option.key] || ''}
                      onChange={(e) =>
                        setTempFilterValues({
                          ...tempFilterValues,
                          [option.key]: e.target.value
                        })
                      }
                      style={{ marginBottom: 8 }}
                    />
                  )}

                  <div className="flex gap-2">
                    <Button
                      type="primary"
                      size="small"
                      style={{ flex: 1 }}
                      onClick={() => applyFilter(option.key, tempFilterValues[option.key] || '')}
                    >
                      Apply
                    </Button>

                    {isActive && (
                      <Button danger size="small" onClick={() => clearFilter(option.key)}>
                        Clear
                      </Button>
                    )}
                  </div>
                </div>
              }
            >
              <Button type={isActive ? 'primary' : 'default'} size="small">
                {isActive ? `${option.label}: ${activeFilter!.value}` : option.label}
              </Button>
            </Popover>
          )
        })}

        {activeFilters.length > 0 && (
          <Button size="small" onClick={clearAllFilters}>
            Clear All
          </Button>
        )}
      </Space>
    )
  }

  const formatDate = (dateString?: string) => {
    if (!dateString) return '-'
    return dayjs(dateString).tz(timezone).format('YYYY-MM-DD HH:mm:ss')
  }

  const columns = [
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: MessageStatus) => (
        <Tag color={statusConfig[status]?.color || 'default'} icon={statusConfig[status]?.icon}>
          {statusConfig[status]?.label || status}
        </Tag>
      )
    },
    {
      title: 'Channel',
      dataIndex: 'channel',
      key: 'channel',
      render: (channel: string) => <Tag>{channel}</Tag>
    },
    {
      title: 'Contact ID',
      dataIndex: 'contact_email',
      key: 'contact_email',
      ellipsis: true
    },
    {
      title: 'Template ID',
      dataIndex: 'template_id',
      key: 'template_id',
      ellipsis: true
    },
    {
      title: 'Sent At',
      dataIndex: 'sent_at',
      key: 'sent_at',
      render: (date: string) => (
        <Tooltip title={formatDate(date)}>{dayjs(date).tz(timezone).fromNow()}</Tooltip>
      )
    },
    {
      title: 'Last Update',
      key: 'last_update',
      render: (_: unknown, record: MessageHistory) => {
        // Find the most recent timestamp
        const timestamps = [
          { label: 'Delivered', date: record.delivered_at },
          { label: 'Failed', date: record.failed_at },
          { label: 'Opened', date: record.opened_at },
          { label: 'Clicked', date: record.clicked_at },
          { label: 'Bounced', date: record.bounced_at },
          { label: 'Complained', date: record.complained_at },
          { label: 'Unsubscribed', date: record.unsubscribed_at }
        ]
          .filter((item) => item.date)
          .sort((a, b) => new Date(b.date!).getTime() - new Date(a.date!).getTime())

        if (timestamps.length === 0) {
          return '-'
        }

        const latest = timestamps[0]
        return (
          <Tooltip title={`${latest.label}: ${formatDate(latest.date)}`}>
            {dayjs(latest.date).tz(timezone).fromNow()}
          </Tooltip>
        )
      }
    },
    {
      title: 'Error',
      dataIndex: 'error',
      key: 'error',
      render: (error?: string) =>
        error ? (
          <Tooltip title={error}>
            <Tag color="red">Error</Tag>
          </Tooltip>
        ) : null
    }
  ]

  if (error) {
    return (
      <div>
        <Title level={4}>Error loading data</Title>
        <Text type="danger">{(error as Error)?.message}</Text>
      </div>
    )
  }

  // Show empty state when there's no data and no loading
  const showEmptyState = !isLoading && (!allMessages || allMessages.length === 0)

  return (
    <div>
      <div className="flex justify-between items-center my-6">{renderFilterButtons()}</div>

      <Table
        dataSource={allMessages}
        columns={columns}
        rowKey="id"
        loading={isLoading && !isLoadingMore}
        pagination={false}
        locale={{
          emptyText: showEmptyState
            ? 'No messages found. Try adjusting your filters.'
            : 'Loading...'
        }}
        expandable={{
          expandedRowRender: (record: MessageHistory) => (
            <div className="px-4">
              <pre className="bg-gray-50 p-4 rounded">
                {JSON.stringify(record.message_data, null, 2)}
              </pre>
            </div>
          )
        }}
        className="border border-gray-200 rounded-md"
      />

      {messagesData?.next_cursor && (
        <div className="flex justify-center mt-4 mb-8">
          <Button onClick={handleLoadMore} loading={isLoadingMore}>
            Load More
          </Button>
        </div>
      )}
    </div>
  )
}

// Webhooks Tab
const WebhooksTab: React.FC<{ workspaceId: string }> = ({ workspaceId }) => {
  const { workspaces } = useAuth()

  // Current workspace details
  const currentWorkspace = workspaces.find((workspace) => workspace.id === workspaceId)

  // Get email provider integrations
  const emailProviders = currentWorkspace?.integrations?.filter(
    (integration) => integration.type === 'email'
  )

  // Fetch webhook status for the first email provider
  const defaultProviderId = emailProviders?.[0]?.id

  const {
    data: webhookStatus,
    isLoading,
    error
  } = useQuery({
    queryKey: ['webhook-status', workspaceId, defaultProviderId],
    queryFn: () =>
      getWebhookStatus({
        workspace_id: workspaceId,
        integration_id: defaultProviderId!
      }),
    enabled: !!workspaceId && !!defaultProviderId
  })

  if (!defaultProviderId) {
    return (
      <div className="text-center py-8">
        <Title level={4}>No Email Provider Configured</Title>
        <Text type="secondary">
          You need to configure an email provider integration to view webhook status.
        </Text>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <Title level={4}>Error loading webhook status</Title>
        <Text type="danger">{(error as Error)?.message}</Text>
      </div>
    )
  }

  const columns = [
    {
      title: 'Event Type',
      dataIndex: 'event_type',
      key: 'event_type',
      render: (eventType: string) => (
        <Tag>{eventType.charAt(0).toUpperCase() + eventType.slice(1)}</Tag>
      )
    },
    {
      title: 'Webhook URL',
      dataIndex: 'url',
      key: 'url',
      render: (url: string) => (
        <Text ellipsis style={{ maxWidth: 400 }}>
          {url}
        </Text>
      )
    },
    {
      title: 'Status',
      dataIndex: 'active',
      key: 'active',
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'red'}>{active ? 'Active' : 'Inactive'}</Tag>
      )
    }
  ]

  return (
    <div>
      <Card className="mb-4">
        <Space direction="vertical">
          <div>
            <Text strong>Provider:</Text>{' '}
            <Tag>{webhookStatus?.status.email_provider_kind || 'Unknown'}</Tag>
          </div>
          <div>
            <Text strong>Status:</Text>{' '}
            <Tag color={webhookStatus?.status.is_registered ? 'green' : 'orange'}>
              {webhookStatus?.status.is_registered ? 'Registered' : 'Not Registered'}
            </Tag>
          </div>
          {webhookStatus?.status.error && (
            <div>
              <Text strong>Error:</Text> <Text type="danger">{webhookStatus.status.error}</Text>
            </div>
          )}
        </Space>
      </Card>

      <Table
        dataSource={webhookStatus?.status.endpoints || []}
        columns={columns}
        rowKey="webhook_id"
        loading={isLoading}
        pagination={false}
        className="border border-gray-200 rounded-md"
      />
    </div>
  )
}

export function LogsPage() {
  const { workspaceId } = useParams({ strict: false })

  if (!workspaceId) {
    return (
      <div>
        <Title level={4}>Workspace Required</Title>
        <Text type="secondary">Please select a workspace to view logs.</Text>
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <div className="text-2xl font-medium">Logs</div>
        <Text type="secondary">Monitor message delivery status and webhook events</Text>
      </div>

      <Tabs
        defaultActiveKey="messages"
        items={[
          {
            key: 'messages',
            label: 'Message History',
            children: <MessagesHistoryTab workspaceId={workspaceId} />
          },
          {
            key: 'webhooks',
            label: 'Webhooks',
            children: <WebhooksTab workspaceId={workspaceId} />
          }
        ]}
      />
    </div>
  )
}
