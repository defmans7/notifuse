import { useState, useEffect } from 'react'
import {
  Button,
  Drawer,
  Form,
  Input,
  Select,
  Space,
  message,
  Tabs,
  Row,
  Col,
  Divider,
  Tag,
  Alert
} from 'antd'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { templatesApi } from '../../services/api/template'
import type { Template } from '../../services/api/types'
import { PlusOutlined } from '@ant-design/icons'
import { useParams } from '@tanstack/react-router'
import { DefaultEditor } from '../../components/email_editor'
import { kebabCase } from 'lodash'
import IphoneEmailPreview from './PhonePreview'

// Extended template interface with additional properties
interface ExtendedTemplate extends Template {
  category?: string
  from_address?: string
  from_name?: string
  reply_to?: string
  visual_editor_tree?: any
  utm_source?: string
  utm_medium?: string
  utm_campaign?: string
}

interface CreateTemplateDrawerProps {
  template?: ExtendedTemplate
  workspaceId?: string
  buttonProps?: any
  onClose?: () => void
  category?: string
  utmSource?: string
  utmMedium?: string
  utmCampaign?: string
}

export function CreateTemplateDrawer({
  template,
  workspaceId: propWorkspaceId,
  buttonProps = {},
  onClose,
  category,
  utmSource,
  utmMedium,
  utmCampaign
}: CreateTemplateDrawerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [form] = Form.useForm()
  const { workspaceId: paramWorkspaceId } = useParams({ from: '/workspace/$workspaceId/templates' })
  const workspaceId = propWorkspaceId || paramWorkspaceId
  const queryClient = useQueryClient()
  const [contentType, setContentType] = useState<'html' | 'plain'>(template?.content_type || 'html')
  const [content, setContent] = useState(template?.content || '')
  const [tab, setTab] = useState<string>('settings')
  const [loading, setLoading] = useState(false)
  const [editorHeight, setEditorHeight] = useState(0)

  // Add Form.useWatch for the email fields
  const fromName = Form.useWatch(['email', 'from_name'], form)
  const emailSubject = Form.useWatch(['email', 'subject'], form)
  const emailContent = Form.useWatch(['email', 'content'], form)

  // Calculate editor height based on drawer dimensions
  useEffect(() => {
    if (isOpen && tab === 'template') {
      const calculateHeight = () => {
        const doc = document.querySelector('.ant-drawer')
        const topbarHeight = 65
        const contentHeight = doc ? parseInt(window.getComputedStyle(doc).height) - topbarHeight : 0
        setEditorHeight(contentHeight)
      }

      calculateHeight()
      window.addEventListener('resize', calculateHeight)

      return () => {
        window.removeEventListener('resize', calculateHeight)
      }
    }
  }, [isOpen, tab])

  // Auto-open drawer when template is provided (for edit mode)
  useEffect(() => {
    if (template) {
      showDrawer()
    }
  }, [template])

  // When content type changes, update the form field
  useEffect(() => {
    if (isOpen) {
      form.setFieldsValue({ content_type: contentType })
    }
  }, [contentType, form, isOpen])

  // When content changes from CodeEditor, update the form field
  useEffect(() => {
    if (isOpen) {
      form.setFieldsValue({ content })
    }
  }, [content, form, isOpen])

  const createTemplateMutation = useMutation({
    mutationFn: (values: any) => {
      if (template) {
        return templatesApi.update({
          ...values,
          workspace_id: workspaceId,
          id: template.id
        })
      } else {
        return templatesApi.create({
          ...values,
          workspace_id: workspaceId
        })
      }
    },
    onSuccess: () => {
      message.success(`Template ${template ? 'updated' : 'created'} successfully`)
      handleClose()
      queryClient.invalidateQueries({ queryKey: ['templates', workspaceId] })
    },
    onError: (error) => {
      message.error(`Failed to ${template ? 'update' : 'create'} template: ${error.message}`)
    }
  })

  const showDrawer = () => {
    if (template) {
      form.setFieldsValue({
        name: template.name,
        id: template.id || kebabCase(template.name),
        description: template.description,
        content: template.content,
        content_type: template.content_type,
        subject: template.subject,
        category: template.category || undefined,
        email: {
          from_address: template.from_address || '',
          from_name: template.from_name || '',
          reply_to: template.reply_to || '',
          subject: template.subject || '',
          content: template.content || '',
          visual_editor_tree: template.visual_editor_tree || null
        },
        utm_source: template.utm_source || utmSource || '',
        utm_medium: template.utm_medium || utmMedium || 'email',
        utm_campaign: template.utm_campaign || utmCampaign || ''
      })
      setContentType(template.content_type)
      setContent(template.content)
    }
    setIsOpen(true)
  }

  const handleClose = () => {
    setIsOpen(false)
    form.resetFields()
    setContent('')
    setContentType('html')
    setTab('settings')
    if (onClose) {
      onClose()
    }
  }

  const goNext = () => {
    setTab('template')
  }

  const handleSubmit = () => {
    setLoading(true)
    form
      .validateFields()
      .then((values) => {
        const urlParams = {
          utm_source: values.utm_source,
          utm_medium: values.utm_medium,
          utm_campaign: values.utm_campaign,
          utm_content: values.id,
          utm_id: '{{ rmd_utm_id }}'
        }

        // Add logic to export HTML if needed
        // const result = ExportHTML(values.email.visual_editor_tree, urlParams)
        // if (result.errors && result.errors.length > 0) {
        //   message.error(result.errors[0].formattedMessage)
        //   setLoading(false)
        //   return
        // }
        // values.email.content = result.html

        // For now, just use the content directly
        values.content = values.email.content || content

        createTemplateMutation.mutate(values)
      })
      .catch((info) => {
        console.log('Validate Failed:', info)
        if (info.errorFields) {
          info.errorFields.forEach((field: any) => {
            if (
              [
                'name',
                'id',
                'email.from_address',
                'email.from_name',
                'email.subject',
                'email.reply_to'
              ].indexOf(field.name[0]) !== -1
            ) {
              setTab('settings')
            }
          })
        }
        setLoading(false)
      })
  }

  return (
    <>
      <Button type="primary" onClick={showDrawer} icon={<PlusOutlined />} {...buttonProps}>
        {template ? 'Edit Template' : 'Create Template'}
      </Button>
      {isOpen && (
        <Drawer
          title={<>{template ? 'Edit email template' : 'Create an email template'}</>}
          closable={true}
          keyboard={false}
          maskClosable={false}
          width={tab === 'settings' ? 960 : '95%'}
          open={isOpen}
          onClose={handleClose}
          className="drawer-no-transition drawer-body-no-padding"
          extra={
            <div className="text-right">
              <Space>
                <Button type="link" loading={loading} onClick={handleClose}>
                  Cancel
                </Button>

                {tab === 'settings' && (
                  <Button type="primary" onClick={goNext}>
                    Next
                  </Button>
                )}
                {tab === 'template' && (
                  <Button type="primary" ghost onClick={() => setTab('settings')}>
                    Previous
                  </Button>
                )}

                {tab === 'template' && (
                  <Button
                    loading={loading || createTemplateMutation.isPending}
                    onClick={handleSubmit}
                    type="primary"
                  >
                    Save
                  </Button>
                )}
              </Space>
            </div>
          }
        >
          <div className="flex justify-center">
            <Tabs
              activeKey={tab}
              centered
              onChange={(k) => setTab(k)}
              style={{ display: 'inline-block' }}
              className="tabs-in-header"
              items={[
                {
                  key: 'settings',
                  label: '1. Settings'
                },
                {
                  key: 'template',
                  label: '2. Template'
                }
              ]}
            />
          </div>
          <div className="relative">
            {tab === 'settings' ? (
              <Form
                form={form}
                layout="vertical"
                initialValues={{
                  content_type: 'html',
                  category: category || undefined,
                  utm_source: utmSource || '',
                  utm_medium: utmMedium || 'email',
                  utm_campaign: utmCampaign || ''
                }}
              >
                <div className="p-8">
                  <Row gutter={24}>
                    <Col span={8}>
                      <Form.Item name="name" label="Template name" rules={[{ required: true }]}>
                        <Input
                          placeholder="i.e: Welcome Email"
                          onChange={(e: any) => {
                            if (!template) {
                              const id = kebabCase(e.target.value)
                              form.setFieldsValue({ id: id })
                            }
                          }}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        name="id"
                        label="Template ID (utm_content)"
                        rules={[
                          {
                            required: true,
                            type: 'string',
                            pattern: /^[a-z0-9]+(-[a-z0-9]+)*$/,
                            message: 'ID must contain only lowercase letters, numbers, and hyphens'
                          }
                        ]}
                      >
                        <Input
                          disabled={template ? true : false}
                          placeholder="i.e: welcome-email"
                        />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        name="category"
                        label="Category"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Select
                          placeholder="Select category"
                          disabled={category ? true : false}
                          options={[
                            {
                              value: 'transactional',
                              label: <Tag color="green">Transactional</Tag>
                            },
                            {
                              value: 'campaign',
                              label: <Tag color="purple">Campaign</Tag>
                            },
                            {
                              value: 'automation',
                              label: <Tag color="cyan">Automation</Tag>
                            },
                            {
                              value: 'other',
                              label: <Tag color="magenta">Other...</Tag>
                            }
                          ]}
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  <div className="text-lg my-8 font-bold">Sender</div>
                  <Row gutter={24}>
                    <Col span={12}>
                      <Row gutter={24}>
                        <Col span={12}>
                          <Form.Item
                            name={['email', 'from_address']}
                            label="Sender email address"
                            rules={[{ required: true, type: 'email' }]}
                          >
                            <Input />
                          </Form.Item>
                        </Col>
                        <Col span={12}>
                          <Form.Item
                            name={['email', 'from_name']}
                            label="Sender name"
                            rules={[{ required: true, type: 'string' }]}
                          >
                            <Input />
                          </Form.Item>
                        </Col>
                      </Row>

                      <Form.Item
                        name={['email', 'subject']}
                        label="Email subject"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Input placeholder="Templating markup allowed" />
                      </Form.Item>
                      <Form.Item
                        name={['email', 'subject_preview']}
                        label="Subject preview"
                        rules={[{ required: true, type: 'string' }]}
                      >
                        <Input placeholder="Templating markup allowed" />
                      </Form.Item>

                      <Form.Item
                        name={['email', 'reply_to']}
                        label="Reply to"
                        rules={[{ required: false, type: 'email' }]}
                      >
                        <Input />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <div className="flex justify-center">
                        <IphoneEmailPreview
                          sender={fromName || 'Sender Name'}
                          subject={emailSubject || 'Email Subject'}
                          previewText={emailContent || 'Preview text will appear here...'}
                          timestamp="Now"
                          currentTime="12:12"
                        />
                      </div>
                    </Col>
                  </Row>

                  <div className="text-lg my-8 font-bold">URL Tracking</div>

                  {utmCampaign && (
                    <Alert
                      type="info"
                      showIcon
                      className="mb-6"
                      message="The utm_source / medium / campaign parameters are already defined at the Campaign level."
                    />
                  )}
                  <Row gutter={24}>
                    <Col span={8}>
                      <Form.Item
                        name="utm_source"
                        label="utm_source"
                        rules={[{ required: false, type: 'string' }]}
                      >
                        <Input placeholder="business.com" disabled={utmSource ? true : false} />
                      </Form.Item>
                    </Col>

                    <Col span={8}>
                      <Form.Item
                        name="utm_medium"
                        label="utm_medium"
                        rules={[{ required: false, type: 'string' }]}
                      >
                        <Input placeholder="email" disabled={utmMedium ? true : false} />
                      </Form.Item>
                    </Col>

                    <Col span={8}>
                      <Form.Item
                        name="utm_campaign"
                        label="utm_campaign"
                        rules={[{ required: false, type: 'string' }]}
                      >
                        <Input disabled={utmCampaign ? true : false} />
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              </Form>
            ) : (
              <div>
                <DefaultEditor
                  initialValue={undefined}
                  height={editorHeight}
                  onChange={(html) => {
                    setContent(html)
                    if (isOpen) {
                      form.setFieldsValue({
                        ['email']: { ...form.getFieldValue('email'), content: html }
                      })
                    }
                  }}
                  onUserBlocksUpdate={async () => {}}
                />
              </div>
            )}
          </div>
        </Drawer>
      )}
    </>
  )
}
