import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Typography, Button, Table, Tooltip, Tag, Space, Popconfirm, message } from 'antd'
import { useParams } from '@tanstack/react-router'
import { templatesApi } from '../services/api/template'
import type { Template, Workspace } from '../services/api/types'
import { EditOutlined, EyeOutlined, DeleteOutlined } from '@ant-design/icons'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'
import { useAuth } from '../contexts/AuthContext'
import dayjs from '../lib/dayjs'
import TemplatePreviewPopover from '../components/templates/TemplatePreviewPopover'

const { Title, Paragraph, Text } = Typography

export function TemplatesPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/templates' })
  const queryClient = useQueryClient()
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

  const deleteMutation = useMutation({
    mutationFn: templatesApi.delete,
    onSuccess: () => {
      message.success('Template deleted successfully')
      queryClient.invalidateQueries({ queryKey: ['templates', workspaceId] })
    },
    onError: (error: any) => {
      const errorMsg = error?.response?.data?.error || error.message
      message.error(`Failed to delete template: ${errorMsg}`)
    }
  })

  const handleDelete = (templateId: string) => {
    deleteMutation.mutate({ workspace_id: workspaceId!, id: templateId })
  }

  const hasTemplates = !isLoading && data?.templates && data.templates.length > 0

  const handleDrawerClose = () => {
    setSelectedTemplate(null)
  }

  const columns = [
    {
      title: 'Template',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Template) => (
        <div>
          <Text strong>{text}</Text>
          <div>
            <Text type="secondary" className="text-xs">
              <Tooltip title={'ID for API: ' + record.id}>{record.id}</Tooltip>
            </Text>
          </div>
        </div>
      )
    },
    {
      title: 'Category',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => {
        const colorMap: Record<string, string> = {
          transactional: 'green',
          campaign: 'purple',
          automation: 'cyan',
          other: 'magenta'
        }
        return <Tag color={colorMap[category] || 'default'}>{category}</Tag>
      }
    },
    {
      title: 'Subject',
      dataIndex: ['email', 'subject'],
      key: 'subject',
      render: (subject: string, record: Template) => (
        <div>
          <Text>{subject}</Text>
          {record.email?.subject_preview && (
            <div>
              <Text type="secondary" className="text-xs">
                {record.email.subject_preview}
              </Text>
            </div>
          )}
        </div>
      )
    },
    {
      title: 'UTM',
      key: 'utm',
      render: (_: any, record: Template) => (
        <div className="space-y-1">
          {record.utm_source && (
            <div>
              <Text type="secondary" className="text-xs">
                utm_source: {record.utm_source}
              </Text>
            </div>
          )}
          {record.utm_medium && (
            <div>
              <Text type="secondary" className="text-xs">
                utm_medium: {record.utm_medium}
              </Text>
            </div>
          )}
          {record.utm_campaign && (
            <div>
              <Text type="secondary" className="text-xs">
                utm_campaign: {record.utm_campaign}
              </Text>
            </div>
          )}
        </div>
      )
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => (
        <Tooltip
          title={
            dayjs(date).tz(workspace?.settings.timezone).format('llll') +
            ' in ' +
            workspace?.settings.timezone
          }
        >
          <span>{dayjs(date).format('ll')}</span>
        </Tooltip>
      )
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Template) => (
        <Space size="middle">
          <TemplatePreviewPopover record={record} workspaceId={workspaceId!}>
            <Button type="text" icon={<EyeOutlined />} />
          </TemplatePreviewPopover>
          <Button type="text" icon={<EditOutlined />} onClick={() => setSelectedTemplate(record)} />
          <Popconfirm
            title="Delete the template?"
            description="Are you sure you want to delete this template? All versions will be deleted."
            onConfirm={() => handleDelete(record.id)}
            okText="Yes, Delete"
            cancelText="Cancel"
            placement="topRight"
          >
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
            />
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <Title level={2}>Templates</Title>
        {workspace && data?.templates && data.templates.length > 0 && (
          <CreateTemplateDrawer workspace={workspace} />
        )}
      </div>

      {isLoading ? (
        <Table columns={columns} dataSource={[]} loading={true} rowKey="id" />
      ) : hasTemplates ? (
        <Table
          columns={columns}
          dataSource={data.templates}
          rowKey="id"
          pagination={{ hideOnSinglePage: true }}
        />
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
