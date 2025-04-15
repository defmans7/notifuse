import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Typography,
  Button,
  Table,
  Tooltip,
  Tag,
  Space,
  Popconfirm,
  message,
  Segmented
} from 'antd'
import { useParams, useSearch, useNavigate } from '@tanstack/react-router'
import { templatesApi } from '../services/api/template'
import type { Template, Workspace } from '../services/api/types'
import { EditOutlined, EyeOutlined, DeleteOutlined } from '@ant-design/icons'
import { CreateTemplateDrawer } from '../components/templates/CreateTemplateDrawer'
import { useAuth } from '../contexts/AuthContext'
import dayjs from '../lib/dayjs'
import TemplatePreviewPopover from '../components/templates/TemplatePreviewPopover'

const { Title, Paragraph, Text } = Typography

// Define search params interface
interface TemplatesSearch {
  category?: string
}

export function TemplatesPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/templates' })
  // Use useSearch to get query params
  const search = useSearch({ from: '/workspace/$workspaceId/templates' }) as TemplatesSearch
  const navigate = useNavigate({ from: '/workspace/$workspaceId/templates' })
  const queryClient = useQueryClient()
  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null)
  const { workspaces } = useAuth()
  const [workspace, setWorkspace] = useState<Workspace | null>(null)
  // Derive selectedCategory from search params, default to 'all'
  const selectedCategory = search.category || 'all'

  // Function to update search params
  const setSelectedCategory = (category: string) => {
    navigate({
      search: (prev) => ({ ...prev, category: category === 'all' ? undefined : category })
    })
  }

  // Backend categories + All
  const categories = [
    { label: 'All', value: 'all' },
    { label: 'Marketing', value: 'marketing' },
    { label: 'Transactional', value: 'transactional' },
    { label: 'Welcome', value: 'welcome' },
    { label: 'Opt-in', value: 'opt_in' },
    { label: 'Unsubscribe', value: 'unsubscribe' },
    { label: 'Bounce', value: 'bounce' },
    { label: 'Blocklist', value: 'blocklist' },
    { label: 'Other', value: 'other' }
  ]

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
    // Use selectedCategory from search params in queryKey
    queryKey: ['templates', workspaceId, selectedCategory],
    queryFn: () => {
      const params: { workspace_id: string; category?: string } = {
        workspace_id: workspaceId
      }
      if (selectedCategory !== 'all') {
        params.category = selectedCategory
      }
      return templatesApi.list(params)
    }
  })

  const deleteMutation = useMutation({
    mutationFn: templatesApi.delete,
    onSuccess: () => {
      message.success('Template deleted successfully')
      // Use selectedCategory from search params in invalidation
      queryClient.invalidateQueries({ queryKey: ['templates', workspaceId, selectedCategory] })
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

      <div className="mb-4">
        <Segmented
          options={categories}
          // Use selectedCategory from search params as value
          value={selectedCategory}
          // Update search params on change
          onChange={(value) => setSelectedCategory(value as string)}
        />
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
          {selectedCategory === 'all' ? (
            <>
              <Title level={4} type="secondary">
                No templates found
              </Title>
              <Paragraph type="secondary">Create your first template to get started</Paragraph>
              <div className="mt-4">
                {workspace && (
                  <CreateTemplateDrawer workspace={workspace} buttonProps={{ size: 'large' }} />
                )}
              </div>
            </>
          ) : (
            <>
              <Title level={4} type="secondary">
                No templates found for category "{selectedCategory}"
              </Title>
              <Paragraph type="secondary">
                Try selecting a different category or{' '}
                <Button type="link" onClick={() => setSelectedCategory('all')} className="p-0">
                  reset the filter
                </Button>
                .
              </Paragraph>
            </>
          )}
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
