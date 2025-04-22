import { useQuery } from '@tanstack/react-query'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Typography,
  Space,
  Tooltip,
  Descriptions,
  Button,
  Divider
} from 'antd'
import { useParams } from '@tanstack/react-router'
import { listsApi } from '../services/api/list'
import { templatesApi } from '../services/api/template'
import type { List, TemplateReference } from '../services/api/types'
import { CreateListDrawer } from '../components/lists/ListDrawer'
import {
  UsergroupAddOutlined,
  CheckCircleOutlined,
  StopOutlined,
  WarningOutlined,
  FrownOutlined,
  EditOutlined
} from '@ant-design/icons'
import { Check, X } from 'lucide-react'
import TemplatePreviewPopover from '../components/templates/TemplatePreviewPopover'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'
import { useAuth } from '../contexts/AuthContext'

const { Title, Paragraph, Text } = Typography

// Component to fetch template data and render the preview popover
const TemplatePreviewButton = ({
  templateRef,
  workspaceId
}: {
  templateRef: TemplateReference
  workspaceId: string
}) => {
  const { workspaces } = useAuth()
  const workspace = workspaces.find((w) => w.id === workspaceId)

  const { data, isLoading } = useQuery({
    queryKey: ['template', workspaceId, templateRef.id, templateRef.version],
    queryFn: async () => {
      const response = await templatesApi.get({
        workspace_id: workspaceId,
        id: templateRef.id,
        version: templateRef.version
      })
      return response.template
    },
    enabled: !!templateRef && !!workspaceId,
    // No need to refetch often - template won't change
    staleTime: 1000 * 60 * 5 // 5 minutes
  })

  if (isLoading || !data) {
    return (
      <Button type="link" size="small" loading={isLoading}>
        preview
      </Button>
    )
  }

  return (
    <Space>
      <TemplatePreviewPopover record={data} workspaceId={workspaceId}>
        <Button type="link" size="small">
          preview
        </Button>
      </TemplatePreviewPopover>
      {workspace && (
        <CreateTemplateDrawer
          template={data}
          workspace={workspace}
          buttonContent="edit"
          buttonProps={{ type: 'link', size: 'small' }}
        />
      )}
    </Space>
  )
}

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
                  {list.is_public && <Tag color="green">Public</Tag>}
                  {!list.is_public && <Tag color="red">Private</Tag>}
                  <CreateListDrawer
                    workspaceId={workspaceId}
                    list={list}
                    buttonProps={{
                      type: 'text',
                      size: 'small',
                      buttonContent: (
                        <Tooltip title="Edit List">
                          <EditOutlined />
                        </Tooltip>
                      )
                    }}
                  />
                </Space>
              }
              bordered={false}
              key={list.id}
            >
              <Row gutter={[16, 16]} wrap={false}>
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

              <Divider />

              <Descriptions size="small" column={2}>
                <Descriptions.Item label="ID">{list.id}</Descriptions.Item>

                <Descriptions.Item label="Description">{list.description}</Descriptions.Item>

                {/* Double Opt-in Template */}
                <Descriptions.Item label="Double Opt-in Template">
                  {list.double_optin_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.double_optin_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>

                {/* Welcome Template */}
                <Descriptions.Item label="Welcome Template">
                  {list.welcome_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.welcome_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>

                {/* Unsubscribe Template */}
                <Descriptions.Item label="Unsubscribe Template">
                  {list.unsubscribe_template ? (
                    <Space>
                      <Check size={16} className="text-green-500" />
                      <TemplatePreviewButton
                        templateRef={list.unsubscribe_template}
                        workspaceId={workspaceId}
                      />
                    </Space>
                  ) : (
                    <X size={16} className="text-red-500" />
                  )}
                </Descriptions.Item>
              </Descriptions>
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
