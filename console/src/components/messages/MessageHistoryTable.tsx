import React, { useEffect, useState, useRef } from 'react'
import { Table, Tag, Tooltip, Button, Spin, Empty, Popover, Badge } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faExclamationTriangle,
  faPaperPlane,
  faCircleCheck,
  faCircleXmark,
  faEye,
  faArrowPointer,
  faBan,
  faTriangleExclamation,
  faCode
} from '@fortawesome/free-solid-svg-icons'
import { faFaceFrown } from '@fortawesome/free-regular-svg-icons'
import dayjs from '../../lib/dayjs'
import { MessageHistory, MessageStatus } from '../../services/api/messages_history'
import { usePrismjs } from '../../components/email_editor/UI/Widgets/PrismJS'

// Separate component for JSON data display
const JsonDataViewer = ({ data }: { data: any }) => {
  const codeRef = useRef<HTMLDivElement>(null)
  // Apply syntax highlighting
  usePrismjs(codeRef, ['line-numbers'])

  const prettyJson = JSON.stringify(data, null, 2)

  return (
    <div
      ref={codeRef}
      className="max-h-96 overflow-auto rounded"
      style={{ maxWidth: '600px', minWidth: '400px' }}
    >
      <pre
        className="line-numbers"
        style={{
          margin: '0',
          borderRadius: '4px',
          padding: 'O',
          fontSize: '12px',
          wordWrap: 'break-word',
          whiteSpace: 'pre-wrap',
          wordBreak: 'normal'
        }}
      >
        <code className="language-json">{prettyJson}</code>
      </pre>
    </div>
  )
}

interface MessageHistoryTableProps {
  messages?: MessageHistory[]
  loading: boolean
  isLoadingMore: boolean
  workspaceTimezone?: string
  nextCursor?: string
  onLoadMore: () => void
  show_email?: boolean
  bordered?: boolean
  size?: 'small' | 'middle' | 'large'
}

export function MessageHistoryTable({
  messages = [],
  loading,
  isLoadingMore,
  workspaceTimezone = 'UTC',
  nextCursor,
  onLoadMore,
  show_email = true,
  bordered = false,
  size = 'small'
}: MessageHistoryTableProps) {
  // Format date using dayjs
  const formatDate = (dateString: string | undefined): string => {
    if (!dateString) return '-'
    return `${dayjs(dateString).format('lll')} in ${workspaceTimezone}`
  }

  // Status badge for message history
  const getStatusBadge = (status: MessageStatus) => {
    const statusConfig: Record<MessageStatus, { color: string; text: string }> = {
      sent: { color: 'blue', text: 'Sent' },
      delivered: { color: 'green', text: 'Delivered' },
      failed: { color: 'red', text: 'Failed' },
      opened: { color: 'cyan', text: 'Opened' },
      clicked: { color: 'geekblue', text: 'Clicked' },
      bounced: { color: 'volcano', text: 'Bounced' },
      complained: { color: 'magenta', text: 'Complained' },
      unsubscribed: { color: 'gold', text: 'Unsubscribed' }
    }

    const config = statusConfig[status] || { color: 'default', text: status }
    return <Badge status={config.color as any} text={config.text} />
  }

  // Define base columns
  const baseColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      render: (id: string) => {
        return (
          <Tooltip title={id}>
            <span className="text-xs text-gray-500">{id.substring(0, 8) + '...'}</span>
          </Tooltip>
        )
      }
    },
    {
      title: 'Template',
      dataIndex: 'template_id',
      key: 'template_id'
    },
    {
      title: 'Broadcast',
      dataIndex: 'broadcast_id',
      key: 'broadcast_id'
    },
    {
      title: 'Status',
      key: 'status',
      render: (record: MessageHistory) => {
        return (
          <div className="flex items-center">
            {getStatusBadge(record.status)}
            {record.error && (
              <Tooltip title={record.error}>
                <FontAwesomeIcon icon={faExclamationTriangle} className="!ml-2" />
              </Tooltip>
            )}
          </div>
        )
      }
    },
    {
      title: 'Events',
      key: 'events',
      render: (record: MessageHistory) => {
        const events = []
        if (record.sent_at)
          events.push(
            <Tooltip key="sent" title={formatDate(record.sent_at)}>
              <Tag bordered={false} color="blue">
                <FontAwesomeIcon icon={faPaperPlane} /> Sent
              </Tag>
            </Tooltip>
          )
        if (record.delivered_at)
          events.push(
            <Tooltip key="delivered" title={formatDate(record.delivered_at)}>
              <Tag bordered={false} color="green">
                <FontAwesomeIcon icon={faCircleCheck} /> Delivered
              </Tag>
            </Tooltip>
          )
        if (record.failed_at)
          events.push(
            <Tooltip key="failed" title={formatDate(record.failed_at)}>
              <Tag bordered={false} color="red">
                <FontAwesomeIcon icon={faCircleXmark} /> Failed
              </Tag>
            </Tooltip>
          )
        if (record.opened_at)
          events.push(
            <Tooltip key="opened" title={formatDate(record.opened_at)}>
              <Tag bordered={false} color="cyan">
                <FontAwesomeIcon icon={faEye} /> Opened
              </Tag>
            </Tooltip>
          )
        if (record.clicked_at)
          events.push(
            <Tooltip key="clicked" title={formatDate(record.clicked_at)}>
              <Tag bordered={false} color="geekblue">
                <FontAwesomeIcon icon={faArrowPointer} /> Clicked
              </Tag>
            </Tooltip>
          )
        if (record.bounced_at)
          events.push(
            <Tooltip key="bounced" title={formatDate(record.bounced_at)}>
              <Tag bordered={false} color="volcano">
                <FontAwesomeIcon icon={faTriangleExclamation} /> Bounced
              </Tag>
            </Tooltip>
          )
        if (record.complained_at)
          events.push(
            <Tooltip key="complained" title={formatDate(record.complained_at)}>
              <Tag bordered={false} color="red">
                <FontAwesomeIcon icon={faFaceFrown} /> Complained
              </Tag>
            </Tooltip>
          )
        if (record.unsubscribed_at)
          events.push(
            <Tooltip key="unsubscribed" title={formatDate(record.unsubscribed_at)}>
              <Tag bordered={false} color="red">
                <FontAwesomeIcon icon={faBan} /> Unsubscribed
              </Tag>
            </Tooltip>
          )
        return <div className="flex items-center gap-1">{events}</div>
      }
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => {
        return <Tooltip title={formatDate(date)}>{dayjs(date).fromNow()}</Tooltip>
      }
    },
    {
      title: 'Data',
      key: 'json_data',
      render: (record: MessageHistory) => {
        const jsonData = record.message_data.data || {}

        return (
          <Popover
            placement="left"
            title="Message Data"
            content={<JsonDataViewer data={jsonData} />}
            trigger="click"
          >
            <Button
              type="text"
              className="opacity-70"
              icon={<FontAwesomeIcon icon={faCode} />}
              title="View JSON data"
            />
          </Popover>
        )
      }
    }
  ]

  // Email column to conditionally add
  const emailColumn = {
    title: 'Contact Email',
    dataIndex: 'contact_email',
    key: 'contact_email',
    render: (email: string) => <span className="text-xs">{email}</span>
  }

  // Build columns array based on show_email prop
  const columns = show_email ? [emailColumn, ...baseColumns] : baseColumns

  if (loading && !isLoadingMore) {
    return (
      <div className="loading-container" style={{ padding: '40px 0', textAlign: 'center' }}>
        <Spin size="large" />
        <div style={{ marginTop: 16 }}>Loading message history...</div>
      </div>
    )
  }

  if (!messages || messages.length === 0) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="No messages found"
        style={{ margin: '40px 0' }}
      />
    )
  }

  return (
    <>
      <Table
        dataSource={messages}
        columns={columns}
        rowKey="id"
        pagination={false}
        size={size}
        className={bordered ? 'border border-gray-300 rounded' : ''}
      />

      {nextCursor && (
        <div className="flex justify-center mt-4 mb-8">
          <Button size="small" onClick={onLoadMore} loading={isLoadingMore}>
            Load More
          </Button>
        </div>
      )}
    </>
  )
}
