import { useQuery } from '@tanstack/react-query'
import { Row, Col, Statistic, Space, Tooltip, Spin } from 'antd'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faPaperPlane,
  faCircleCheck,
  faEye,
  faCircleXmark,
  faFaceFrown
} from '@fortawesome/free-regular-svg-icons'
import { faArrowPointer, faTriangleExclamation, faBan } from '@fortawesome/free-solid-svg-icons'
import { getBroadcastStats } from '../../services/api/messages_history'

interface BroadcastStatsProps {
  workspaceId: string
  broadcastId: string
}

export function BroadcastStats({ workspaceId, broadcastId }: BroadcastStatsProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['broadcast-stats', workspaceId, broadcastId],
    queryFn: async () => {
      return getBroadcastStats(workspaceId, broadcastId)
    },
    // Refetch every minute to keep stats up to date
    refetchInterval: 60000
  })

  const stats = data?.stats || {
    total_sent: 0,
    total_delivered: 0,
    total_opened: 0,
    total_clicked: 0,
    total_failed: 0,
    total_bounced: 0,
    total_complained: 0,
    total_unsubscribed: 0
  }

  const getRate = (numerator: number, denominator: number) => {
    if (denominator === 0) return '-'
    return `${((numerator / denominator) * 100).toFixed(1)}%`
  }

  // Formatter function for statistics that handles loading state
  const formatStat = (value: number | string) => {
    if (isLoading) {
      return <Spin size="small" />
    }
    return value
  }

  return (
    <Row gutter={[16, 16]} wrap className="flex-nowrap overflow-x-auto">
      <Col span={3}>
        <Tooltip title={`${stats.total_sent} total emails sent`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faPaperPlane}
                    style={{ opacity: 0.7 }}
                    className="text-blue-500"
                  />{' '}
                  Sent
                </Space>
              }
              value={stats.total_sent}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_delivered} emails successfully delivered`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faCircleCheck}
                    style={{ opacity: 0.7 }}
                    className="text-green-500"
                  />{' '}
                  Delivered
                </Space>
              }
              value={getRate(stats.total_delivered, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_opened} total opens`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faEye}
                    style={{ opacity: 0.7 }}
                    className="text-purple-500"
                  />{' '}
                  Opens
                </Space>
              }
              value={getRate(stats.total_opened, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_clicked} total clicks`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faArrowPointer}
                    style={{ opacity: 0.7 }}
                    className="text-cyan-500 mr-1"
                  />{' '}
                  Clicks
                </Space>
              }
              value={getRate(stats.total_clicked, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_failed} emails failed to send`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faCircleXmark}
                    style={{ opacity: 0.7 }}
                    className="text-orange-500"
                  />{' '}
                  Failed
                </Space>
              }
              value={getRate(stats.total_failed, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_bounced} emails bounced back`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faTriangleExclamation}
                    style={{ opacity: 0.7 }}
                    className="text-orange-500"
                  />{' '}
                  Bounced
                </Space>
              }
              value={getRate(stats.total_bounced, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_complained} total complaints`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faFaceFrown}
                    style={{ opacity: 0.7 }}
                    className="text-orange-500"
                  />{' '}
                  Complaints
                </Space>
              }
              value={getRate(stats.total_complained, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
      <Col span={3}>
        <Tooltip title={`${stats.total_unsubscribed} total unsubscribes`}>
          <div>
            <Statistic
              title={
                <Space className="font-medium">
                  <FontAwesomeIcon
                    icon={faBan}
                    style={{ opacity: 0.7 }}
                    className="text-orange-500"
                  />{' '}
                  Unsub.
                </Space>
              }
              value={getRate(stats.total_unsubscribed, stats.total_sent)}
              valueStyle={{ fontSize: '16px' }}
              formatter={formatStat}
            />
          </div>
        </Tooltip>
      </Col>
    </Row>
  )
}
