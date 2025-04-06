import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, Row, Col, Typography, Button } from 'antd'
import { useParams } from '@tanstack/react-router'
import { templatesApi } from '../services/api/template'
import type { Template, Workspace } from '../services/api/types'
import { EditOutlined } from '@ant-design/icons'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'
import { useAuth } from '../contexts/AuthContext'

const { Title, Paragraph, Text } = Typography

export function TemplatesPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/templates' })
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)
  const { workspaces } = useAuth()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)

  // current workspace from workspaceId
  useEffect(() => {
    if (workspaces.length > 0) {
      const currentWorkspace = workspaces.find((w) => w.id === workspaceId)
      if (currentWorkspace) {
        setWorkspace(currentWorkspace)
      }
    }
  }, [workspaces, workspaceId])

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
        {workspace && data?.templates && data.templates.length > 0 && (
          <CreateTemplateDrawer workspace={workspace} />
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
      ) : hasTemplates ? (
        <Row gutter={[16, 16]}>
          {data.templates.map((template: Template) => (
            <Col xs={24} sm={12} lg={8} key={template.id}>
              <Card
                title={
                  <div className="flex items-center justify-between">
                    <Text strong>{template.name}</Text>
                  </div>
                }
                variant="outlined"
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

                <div className="mb-4">
                  <Text strong>Subject: </Text>
                  <Text>{template.email?.subject}</Text>
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
            {workspace && (
              <CreateTemplateDrawer workspace={workspace} buttonProps={{ size: 'large' }} />
            )}
          </div>
        </div>
      )}

      {workspace && selectedTemplate && (
        <CreateTemplateDrawer
          template={selectedTemplate}
          workspace={workspace}
          buttonProps={{ style: { display: 'none' } }}
          onClose={handleDrawerClose}
        />
      )}
    </div>
  )
}
