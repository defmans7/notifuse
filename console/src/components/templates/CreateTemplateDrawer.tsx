import { useState, useEffect } from 'react'
import { Button, Drawer, Form, Input, Select, Space, App, Tabs, Row, Col, Tag, Alert } from 'antd'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { templatesApi } from '../../services/api/template'
import type { Template, Workspace, FileManagerSettings } from '../../services/api/types'
import { Editor } from '../../components/email_editor'
import { cloneDeep, kebabCase } from 'lodash'
import IphoneEmailPreview from './PhonePreview'
import { DesktopWidth, Layout } from '../../components/email_editor/UI/Layout'
import { SelectedBlockButtonsProp } from '../../components/email_editor/Editor'
import SelectedBlockButtons from '../../components/email_editor/UI/SelectedBlockButtons'
import ButtonBlockDefinition from '../../components/email_editor/UI/definitions/Button'
import ColumnBlockDefinition from '../../components/email_editor/UI/definitions/Column'
import Columns168BlockDefinition from '../../components/email_editor/UI/definitions/Columns168'
import Columns204BlockDefinition from '../../components/email_editor/UI/definitions/Columns204'
import Columns420BlockDefinition from '../../components/email_editor/UI/definitions/Columns420'
import Columns816BlockDefinition from '../../components/email_editor/UI/definitions/Columns816'
import Columns888BlockDefinition from '../../components/email_editor/UI/definitions/Columns888'
import Columns1212BlockDefinition from '../../components/email_editor/UI/definitions/Columns1212'
import Columns6666BlockDefinition from '../../components/email_editor/UI/definitions/Columns6666'
import DividerBlockDefinition from '../../components/email_editor/UI/definitions/Divider'
import HeadingBlockDefinition from '../../components/email_editor/UI/definitions/Heading'
import ImageBlockDefinition from '../../components/email_editor/UI/definitions/Image'
import OneColumnBlockDefinition from '../../components/email_editor/UI/definitions/OneColumn'
import RootBlockDefinition from '../../components/email_editor/UI/definitions/Root'
import TextBlockDefinition from '../../components/email_editor/UI/definitions/Text'
import LiquidTemplateBlockDefinition from '../../components/email_editor/UI/definitions/Liquid'
import { BlockDefinitionInterface, BlockInterface } from '../../components/email_editor/Block'
import uuid from 'short-uuid'
import { useAuth } from '../../contexts/AuthContext'
import { workspaceService } from '../../services/api/workspace'

interface CreateTemplateDrawerProps {
  workspace: Workspace
  template?: Template
  fromTemplate?: Template
  buttonProps?: any
  buttonContent?: React.ReactNode
  onClose?: () => void
  category?: string
  utmDisabled?: boolean
}

// Combine default block definitions with any custom ones
const blockDefinitions = {
  root: RootBlockDefinition,
  column: ColumnBlockDefinition,
  oneColumn: OneColumnBlockDefinition,
  columns168: Columns168BlockDefinition,
  columns204: Columns204BlockDefinition,
  columns420: Columns420BlockDefinition,
  columns816: Columns816BlockDefinition,
  columns888: Columns888BlockDefinition,
  columns1212: Columns1212BlockDefinition,
  columns6666: Columns6666BlockDefinition,
  image: ImageBlockDefinition,
  divider: DividerBlockDefinition,
  button: ButtonBlockDefinition,
  text: TextBlockDefinition,
  heading: HeadingBlockDefinition,
  liquid: LiquidTemplateBlockDefinition
}

// Helper function to generate a block from definition
const generateBlockFromDefinition = (blockDefinition: BlockDefinitionInterface) => {
  const id = uuid.generate()

  const block: BlockInterface = {
    id: id,
    kind: blockDefinition.kind,
    path: '', // path is set when rendering
    children: blockDefinition.children
      ? blockDefinition.children.map((child: BlockDefinitionInterface) => {
          return generateBlockFromDefinition(child)
        })
      : [],
    data: cloneDeep(blockDefinition.defaultData)
  }

  return block
}
// Create default blocks
const createDefaultBlocks = () => {
  // Create default content blocks
  const text = generateBlockFromDefinition(TextBlockDefinition)
  const heading = generateBlockFromDefinition(HeadingBlockDefinition)
  const logo = generateBlockFromDefinition(ImageBlockDefinition)
  const divider = generateBlockFromDefinition(DividerBlockDefinition)
  const btn = generateBlockFromDefinition(ButtonBlockDefinition)
  const column = generateBlockFromDefinition(OneColumnBlockDefinition)

  // Configure logo
  logo.data.image.src = 'https://notifuse.com/images/logo.png'
  logo.data.image.alt = 'Logo'
  logo.data.image.href = 'https://notifuse.com'
  logo.data.image.width = '100px'

  // Configure heading
  heading.data.paddingControl = 'separate'
  heading.data.paddingTop = '40px'
  heading.data.paddingBottom = '40px'
  heading.data.editorData[0].children[0].text =
    'Hello {{ contact.first_name | default:"there" }} 👋'

  // Configure divider
  divider.data.paddingControl = 'separate'
  divider.data.paddingTop = '40px'
  divider.data.paddingBottom = '20px'
  divider.data.paddingLeft = '200px'
  divider.data.paddingRight = '200px'

  // Configure text
  text.data.editorData[0].children[0].text = 'Welcome to the email editor!'

  // Configure button
  btn.data.button.backgroundColor = '#4e6cff'
  btn.data.button.text = '👉 Click me'

  // Add all blocks to column
  column.children[0].children.push(logo)
  column.children[0].children.push(heading)
  column.children[0].children.push(text)
  column.children[0].children.push(divider)
  column.children[0].children.push(btn)

  // Create root block with column as child
  const rootData = cloneDeep(RootBlockDefinition.defaultData)
  const rootBlock: BlockInterface = {
    id: 'root',
    kind: 'root',
    path: '',
    children: [column],
    data: rootData
  }

  return rootBlock
}

/**
 * Renders a Tag component with the appropriate color for an email template category
 */
export const renderCategoryTag = (category: string) => {
  let color = 'default'

  if (['marketing', 'transactional'].includes(category)) {
    color = 'green'
  } else if (category === 'welcome') {
    color = 'blue'
  } else if (['opt_in', 'unsubscribe', 'bounce', 'blocklist'].includes(category)) {
    color = 'purple'
  }

  return (
    <Tag color={color}>
      {category.charAt(0).toUpperCase() + category.slice(1).replace('_', '-')}
    </Tag>
  )
}

export function CreateTemplateDrawer({
  workspace,
  template,
  fromTemplate,
  buttonProps = {},
  buttonContent,
  onClose,
  category,
  utmDisabled = false
}: CreateTemplateDrawerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [form] = Form.useForm()
  const { refreshWorkspaces } = useAuth()
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<string>('settings')
  const [loading, setLoading] = useState(false)
  const [editorHeight, setEditorHeight] = useState(0)
  const { message } = App.useApp()

  // set the tree apart to avoid rerendering the Email Editor when the tree changes
  const [visualEditorTree, setVisualEditorTree] = useState<BlockInterface>(() => {
    if (template && template.email?.visual_editor_tree) {
      // Check if visual_editor_tree is already an object
      if (typeof template.email.visual_editor_tree === 'object') {
        return template.email.visual_editor_tree as unknown as BlockInterface
      }

      // Otherwise parse it from string
      try {
        return JSON.parse(template.email.visual_editor_tree) as BlockInterface
      } catch (error) {
        console.error('Error parsing visual editor tree:', error)
        message.error('Error loading template: Invalid template data')
        return createDefaultBlocks()
      }
    }
    return createDefaultBlocks()
  })

  // Add Form.useWatch for the email fields
  const fromName = Form.useWatch(['email', 'from_name'], form)
  const emailSubject = Form.useWatch(['email', 'subject'], form)
  const emailPreview = Form.useWatch(['email', 'subject_preview'], form)

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

  const createTemplateMutation = useMutation({
    mutationFn: (values: any) => {
      if (template) {
        return templatesApi.update({
          ...values,
          channel: 'email',
          workspace_id: workspace.id,
          id: template.id
        })
      } else {
        return templatesApi.create({
          ...values,
          channel: 'email',
          workspace_id: workspace.id
        })
      }
    },
    onSuccess: () => {
      message.success(`Template ${template ? 'updated' : 'created'} successfully`)
      handleClose()
      queryClient.invalidateQueries({ queryKey: ['templates', workspace.id] })
      setLoading(false)
    },
    onError: (error) => {
      message.error(`Failed to ${template ? 'update' : 'create'} template: ${error.message}`)
      setLoading(false)
    }
  })

  const defaultTestData = {
    contact: {
      first_name: 'John',
      last_name: 'Doe',
      email: 'john.doe@example.com'
    },
    unsubscribe_link: `${window.API_ENDPOINT}/unsubscribe?email={{ contact.email }}&list_id={{ list.id }}&hmac={{ contact.hmac }}`
  }

  const showDrawer = () => {
    if (template) {
      form.setFieldsValue({
        name: template.name,
        id: template.id || kebabCase(template.name),
        category: template.category || undefined,
        email: {
          from_address: template.email?.from_address || '',
          from_name: template.email?.from_name || '',
          reply_to: template.email?.reply_to || undefined,
          subject: template.email?.subject || '',
          subject_preview: template.email?.subject_preview || '',
          content: template.email?.mjml || '',
          visual_editor_tree: template.email?.visual_editor_tree || createDefaultBlocks()
        },
        test_data: template.test_data || defaultTestData,
        utm_source: template.utm_source || undefined,
        utm_medium: template.utm_medium || 'email',
        utm_campaign: template.utm_campaign || undefined
      })
    } else if (fromTemplate) {
      // Clone template functionality - append "copy" as suffix instead of "Copy of" prefix
      form.setFieldsValue({
        name: `${fromTemplate.name} copy`,
        id: kebabCase(`${fromTemplate.name}-copy`),
        category: fromTemplate.category || category || undefined,
        email: {
          from_address: fromTemplate.email?.from_address || '',
          from_name: fromTemplate.email?.from_name || '',
          reply_to: fromTemplate.email?.reply_to || undefined,
          subject: fromTemplate.email?.subject || '',
          subject_preview: fromTemplate.email?.subject_preview || '',
          content: fromTemplate.email?.mjml || '',
          visual_editor_tree: fromTemplate.email?.visual_editor_tree || createDefaultBlocks()
        },
        test_data: fromTemplate.test_data || defaultTestData,
        utm_source: fromTemplate.utm_source || undefined,
        utm_medium: fromTemplate.utm_medium || 'email',
        utm_campaign: fromTemplate.utm_campaign || undefined
      })

      // Update the visual editor tree
      if (fromTemplate.email?.visual_editor_tree) {
        if (typeof fromTemplate.email.visual_editor_tree === 'object') {
          setVisualEditorTree(fromTemplate.email.visual_editor_tree as unknown as BlockInterface)
        } else {
          try {
            setVisualEditorTree(JSON.parse(fromTemplate.email.visual_editor_tree) as BlockInterface)
          } catch (error) {
            console.error('Error parsing visual editor tree:', error)
            message.error('Error loading template: Invalid template data')
          }
        }
      }
    }
    setIsOpen(true)
  }

  const handleClose = () => {
    setIsOpen(false)
    form.resetFields()
    setTab('settings')
    if (onClose) {
      onClose()
    }
  }

  const goNext = () => {
    setTab('template')
  }

  // Function to handle workspace settings update
  const handleUpdateWorkspaceSettings = async (settings: FileManagerSettings): Promise<void> => {
    try {
      // Update workspace using workspace service
      await workspaceService.update({
        id: workspace.id,
        name: workspace.name,
        settings: {
          ...workspace.settings,
          file_manager: settings
        }
      })

      // Refresh workspaces from context
      await refreshWorkspaces()

      message.success('Workspace settings updated successfully')
    } catch (error: any) {
      console.error('Error updating workspace settings:', error)
      message.error(`Failed to update workspace settings: ${error.message}`)
    }
  }

  return (
    <>
      <Button type="primary" onClick={showDrawer} {...buttonProps}>
        {buttonContent ||
          (template ? 'Edit Template' : fromTemplate ? 'Clone Template' : 'Create Template')}
      </Button>
      {isOpen && (
        <Drawer
          title={
            <>
              {template
                ? 'Edit email template'
                : fromTemplate
                  ? 'Clone email template'
                  : 'Create an email template'}
            </>
          }
          closable={true}
          keyboard={false}
          maskClosable={false}
          width={'95%'}
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
                    onClick={() => {
                      form.submit()
                    }}
                    type="primary"
                  >
                    Save
                  </Button>
                )}
              </Space>
            </div>
          }
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={(values) => {
              setLoading(true)
              values.email.visual_editor_tree = visualEditorTree
              createTemplateMutation.mutate(values)
            }}
            onFinishFailed={(info) => {
              if (info.errorFields) {
                info.errorFields.forEach((field: any) => {
                  // field.name can be an array, so we need to concatenate the array into a string
                  const fieldName = field.name.join('.')
                  if (
                    [
                      'name',
                      'id',
                      'category',
                      'email.from_address',
                      'email.from_name',
                      'email.subject',
                      'email.subject_preview',
                      'email.reply_to'
                    ].indexOf(fieldName) !== -1
                  ) {
                    setTab('settings')
                  }
                })
              }
              setLoading(false)
            }}
            initialValues={{
              'email.visual_editor_tree': visualEditorTree,
              category: category || undefined,
              utm_source: '',
              utm_medium: 'email',
              utm_campaign: '',
              test_data: defaultTestData
            }}
          >
            <Form.Item name="test_data" hidden>
              <Input type="hidden" />
            </Form.Item>

            <div className="flex justify-center">
              <Tabs
                activeKey={tab}
                centered
                onChange={(k) => setTab(k)}
                style={{ display: 'inline-block' }}
                className="tabs-in-header"
                destroyInactiveTabPane={false}
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
              <div style={{ display: tab === 'settings' ? 'block' : 'none' }}>
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
                              form.validateFields(['id'])
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
                          },
                          {
                            validator: async (rule, value) => {
                              if (value && !template) {
                                try {
                                  await templatesApi.get({ workspace_id: workspace.id, id: value })
                                  return Promise.reject('Template ID already exists')
                                } catch (error) {
                                  return Promise.resolve()
                                }
                              }
                              return Promise.resolve()
                            }
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
                              value: 'marketing',
                              label: renderCategoryTag('marketing')
                            },
                            {
                              value: 'transactional',
                              label: renderCategoryTag('transactional')
                            },
                            {
                              value: 'welcome',
                              label: renderCategoryTag('welcome')
                            },
                            {
                              value: 'opt_in',
                              label: renderCategoryTag('opt_in')
                            },
                            {
                              value: 'unsubscribe',
                              label: renderCategoryTag('unsubscribe')
                            },
                            {
                              value: 'bounce',
                              label: renderCategoryTag('bounce')
                            },
                            {
                              value: 'blocklist',
                              label: renderCategoryTag('blocklist')
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
                          previewText={emailPreview || 'Preview text will appear here...'}
                          timestamp="Now"
                          currentTime="12:12"
                        />
                      </div>
                    </Col>
                  </Row>

                  {!utmDisabled && (
                    <>
                      <div className="text-lg mt-4 mb-8 font-bold">URL Tracking</div>

                      <Alert
                        type="info"
                        className="!mb-6"
                        message="The utm parameters will be automatically added to your email links."
                      />

                      <Row gutter={24}>
                        <Col span={8}>
                          <Form.Item
                            name="utm_source"
                            label="utm_source"
                            rules={[{ required: false, type: 'string' }]}
                          >
                            <Input placeholder="business.com" />
                          </Form.Item>
                        </Col>

                        <Col span={8}>
                          <Form.Item
                            name="utm_medium"
                            label="utm_medium"
                            rules={[{ required: false, type: 'string' }]}
                          >
                            <Input placeholder="email" />
                          </Form.Item>
                        </Col>

                        <Col span={8}>
                          <Form.Item
                            name="utm_campaign"
                            label="utm_campaign"
                            rules={[{ required: false, type: 'string' }]}
                          >
                            <Input />
                          </Form.Item>
                        </Col>
                      </Row>
                    </>
                  )}
                </div>
              </div>

              <div style={{ display: tab === 'template' ? 'block' : 'none' }}>
                <Form.Item dependencies={['utm_source', 'utm_medium', 'utm_campaign', 'id']}>
                  {(form) => {
                    const utmSourceValue = form.getFieldValue('utm_source')
                    const utmMediumValue = form.getFieldValue('utm_medium')
                    const utmCampaignValue = form.getFieldValue('utm_campaign')
                    const templateID = form.getFieldValue('id')
                    const testData = form.getFieldValue('test_data')

                    return (
                      <Editor
                        blockDefinitions={blockDefinitions}
                        userBlocks={[]}
                        onUserBlocksUpdate={async () => {}}
                        selectedBlockId={'root'}
                        value={visualEditorTree}
                        onChange={setVisualEditorTree}
                        renderSelectedBlockButtons={(props: SelectedBlockButtonsProp) => (
                          <SelectedBlockButtons {...props} />
                        )}
                        deviceWidth={DesktopWidth}
                        urlParams={{
                          utm_source: utmSourceValue || '',
                          utm_medium: utmMediumValue || 'email',
                          utm_campaign: utmCampaignValue || '',
                          utm_content: templateID || '',
                          utm_id: '{{notifuse_utm_id}}'
                        }}
                        fileManagerSettings={workspace?.settings.file_manager}
                        onUpdateFileManagerSettings={handleUpdateWorkspaceSettings}
                        templateDataValue={JSON.stringify(testData, null, 2)}
                        onUpdateTemplateData={async (templateData: string) => {
                          try {
                            // Check if templateData is already an object
                            const parsedData =
                              typeof templateData === 'object'
                                ? templateData
                                : JSON.parse(templateData)

                            form.setFieldsValue({
                              test_data: parsedData
                            })
                            // Handle template data updates
                            return Promise.resolve()
                          } catch (error) {
                            console.error('Error parsing template data:', error)
                            message.error('Invalid JSON in template data')
                            return Promise.reject(error)
                          }
                        }}
                      >
                        <Layout
                          workspaceId={workspace.id}
                          onUpdateMacro={async (macroId: string) => {
                            console.log('macroId', macroId)
                          }}
                          macros={[]}
                          height={editorHeight}
                        />
                      </Editor>
                    )
                  }}
                </Form.Item>
              </div>
            </div>
          </Form>
        </Drawer>
      )}
    </>
  )
}
