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
        <Row gutter={[16, 16]}>
          {data.lists.map((list: List) => (
            <Col xs={24} sm={12} lg={8} key={list.id}>
              <Card
                title={
                  <div className="flex items-center justify-between">
                    <Text strong>{list.name}</Text>
                    <Space>
                      {list.is_double_optin && <Tag color="green">Double Opt-in</Tag>}
                      {list.is_public && <Tag color="blue">Public</Tag>}
                    </Space>
                  </div>
                }
                bordered={false}
                className="h-full"
              >
                <div className="mb-4">
                  <Text type="secondary">ID: {list.id}</Text>
                </div>

                {list.description && (
                  <Paragraph
                    ellipsis={{ rows: 2, expandable: true, symbol: 'more' }}
                    className="mb-4"
                  >
                    {list.description}
                  </Paragraph>
                )}

                <Row gutter={[16, 16]} className="mt-4">
                  <Col span={8}>
                    <Statistic
                      title="Active"
                      value={list.total_active}
                      prefix={<CheckCircleOutlined className="text-green-500" />}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="Pending"
                      value={list.total_pending}
                      prefix={<UsergroupAddOutlined className="text-blue-500" />}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                  <Col span={8}>
                    <Statistic
                      title="Unsub"
                      value={list.total_unsubscribed}
                      prefix={<StopOutlined className="text-gray-500" />}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                </Row>
                <Row gutter={[16, 16]} className="mt-2">
                  <Col span={12}>
                    <Statistic
                      title="Bounced"
                      value={list.total_bounced}
                      prefix={<WarningOutlined className="text-yellow-500" />}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic
                      title="Complaints"
                      value={list.total_complained}
                      prefix={<FrownOutlined className="text-red-500" />}
                      valueStyle={{ fontSize: '16px' }}
                    />
                  </Col>
                </Row>
              </Card>
            </Col>
          ))}
        </Row>
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
