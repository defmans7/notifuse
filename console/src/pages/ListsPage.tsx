import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, Row, Col, Statistic, Tag, Skeleton, Typography, Space } from 'antd'
import { useParams } from '@tanstack/react-router'
import { listsApi } from '../services/api/list'
import type { List } from '../services/api/types'
import { CreateListDrawer } from '../components/lists/CreateListDrawer'
import {
  UsergroupAddOutlined,
  CheckCircleOutlined,
  StopOutlined,
  WarningOutlined,
  FrownOutlined
} from '@ant-design/icons'

const { Title, Paragraph, Text } = Typography

export function ListsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/lists' })

  const { data, isLoading } = useQuery({
    queryKey: ['lists', workspaceId],
    queryFn: () => {
      return listsApi.list({ workspace_id: workspaceId })
    }
  })

  const hasLists = !isLoading && data?.lists && data.lists.length > 0

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2}>Lists</Title>
        {(isLoading || hasLists) && <CreateListDrawer workspaceId={workspaceId} />}
      </div>

      {isLoading ? (
        <Row gutter={[16, 16]}>
          {[1, 2, 3].map((key) => (
            <Col xs={24} sm={12} lg={8} key={key}>
              <Card loading variant="outlined" />
            </Col>
          ))}
        </Row>
      ) : hasLists ? (
        <div className="space-y-4">
          {data.lists.map((list: List) => (
            <Card
              title={
                <div className="flex items-center justify-between">
                  <Text strong>{list.name}</Text>
                </div>
              }
              extra={
                <Space>
                  {list.is_double_optin && <Tag color="green">Double Opt-in</Tag>}
                  {list.is_public && <Tag color="blue">Public</Tag>}
                  <Text type="secondary">ID: {list.id}</Text>
                </Space>
              }
              bordered={false}
              key={list.id}
            >
              <div className="mb-4">
                {list.description && (
                  <Paragraph
                    ellipsis={{ rows: 1, expandable: true, symbol: 'more' }}
                    className="flex-1 mb-0"
                  >
                    {list.description}
                  </Paragraph>
                )}
              </div>

              <Row gutter={[16, 16]} className="mt-4" wrap={false}>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <CheckCircleOutlined className="text-green-500" /> Active
                      </Space>
                    }
                    value={list.total_active}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <UsergroupAddOutlined className="text-blue-500" /> Pending
                      </Space>
                    }
                    value={list.total_pending}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <StopOutlined className="text-gray-500" /> Unsub
                      </Space>
                    }
                    value={list.total_unsubscribed}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <WarningOutlined className="text-yellow-500" /> Bounced
                      </Space>
                    }
                    value={list.total_bounced}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
                <Col flex="1">
                  <Statistic
                    title={
                      <Space>
                        <FrownOutlined className="text-red-500" /> Complaints
                      </Space>
                    }
                    value={list.total_complained}
                    valueStyle={{ fontSize: '16px' }}
                  />
                </Col>
              </Row>
            </Card>
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <Title level={4} type="secondary">
            No lists found
          </Title>
          <Paragraph type="secondary">Create your first list to get started</Paragraph>
          <div className="mt-4">
            <CreateListDrawer workspaceId={workspaceId} buttonProps={{ size: 'large' }} />
          </div>
        </div>
      )}
    </div>
  )
}
