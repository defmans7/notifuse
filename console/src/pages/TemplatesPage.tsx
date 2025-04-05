import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, Row, Col, Typography, Space, Tag, Button } from 'antd'
import { useParams } from '@tanstack/react-router'
import { templatesApi } from '../services/api/template'
import type { Template } from '../services/api/types'
import { FileTextOutlined, Html5Outlined, EditOutlined } from '@ant-design/icons'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'

const { Title, Paragraph, Text } = Typography

export function TemplatesPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/templates' })
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)

  const { data, isLoading } = useQuery({
    queryKey: ['templates', workspaceId],
    queryFn: () => {
      return templatesApi.list({ workspace_id: workspaceId })
    }
  })

  const hasTemplates = !isLoading && data?.templates && data.templates.length > 0

  const handleDrawerClose = () => {
    setSelectedTemplate(null)
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2}>Templates</Title>
        <CreateTemplateDrawer workspaceId={workspaceId} />
      </div>

      {isLoading ? (
        <Row gutter={[16, 16]}>
          {[1, 2, 3].map((key) => (
            <Col xs={24} sm={12} lg={8} key={key}>
              <Card loading variant="outlined" />
            </Col>
          ))}
        </Row>
      ) : hasTemplates ? (
        <Row gutter={[16, 16]}>
          {data.templates.map((template: Template) => (
            <Col xs={24} sm={12} lg={8} key={template.id}>
              <Card
                title={
                  <div className="flex items-center justify-between">
                    <Text strong>{template.name}</Text>
                    <Space>
                      <Tag color={template.content_type === 'html' ? 'blue' : 'green'}>
                        {template.content_type === 'html' ? (
                          <>
                            <Html5Outlined /> HTML
                          </>
                        ) : (
                          <>
                            <FileTextOutlined /> Plain Text
                          </>
                        )}
                      </Tag>
                    </Space>
                  </div>
                }
                bordered={false}
                className="h-full"
                extra={
                  <Button
                    type="text"
                    icon={<EditOutlined />}
                    onClick={() => setSelectedTemplate(template)}
                  />
                }
              >
                <div className="mb-4">
                  <Text type="secondary">ID: {template.id}</Text>
                </div>

                {template.description && (
                  <Paragraph
                    ellipsis={{ rows: 2, expandable: true, symbol: 'more' }}
                    className="mb-4"
                  >
                    {template.description}
                  </Paragraph>
                )}

                <div className="mb-4">
                  <Text strong>Subject: </Text>
                  <Text>{template.subject}</Text>
                </div>

                <div className="text-xs text-gray-500 mt-4">
                  <div>Created: {new Date(template.created_at).toLocaleString()}</div>
                  <div>Updated: {new Date(template.updated_at).toLocaleString()}</div>
                </div>
              </Card>
            </Col>
          ))}
        </Row>
      ) : (
        <div className="text-center py-12">
          <Title level={4} type="secondary">
            No templates found
          </Title>
          <Paragraph type="secondary">Create your first template to get started</Paragraph>
          <div className="mt-4">
            <CreateTemplateDrawer workspaceId={workspaceId} buttonProps={{ size: 'large' }} />
          </div>
        </div>
      )}

      {selectedTemplate && (
        <CreateTemplateDrawer
          template={selectedTemplate}
          workspaceId={workspaceId}
          buttonProps={{ style: { display: 'none' } }}
          onClose={handleDrawerClose}
        />
      )}
    </div>
  )
}
