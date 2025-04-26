import { useState, useEffect } from 'react'
import {
  Button,
  Drawer,
  Form,
  Input,
  Select,
  Space,
  App,
  Tabs,
  Row,
  Col,
  Switch,
  DatePicker,
  message,
  InputNumber,
  Popconfirm
} from 'antd'
import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query'
import {
  broadcastApi,
  Broadcast,
  BroadcastStatus,
  CreateBroadcastRequest,
  UpdateBroadcastRequest,
  AudienceSettings,
  ScheduleSettings,
  BroadcastTestSettings,
  BroadcastVariation
} from '../../services/api/broadcast'
import { templatesApi } from '../../services/api/template'
import type { Workspace } from '../../services/api/types'
import { useParams } from '@tanstack/react-router'
import dayjs from '../../lib/dayjs'
import TemplateSelectorInput from '../templates/TemplateSelectorInput'
import { DeleteOutlined } from '@ant-design/icons'
import React from 'react'

const { TextArea } = Input

// Custom component to handle A/B testing configuration
const ABTestingConfig = ({ form }: { form: any }) => {
  const autoSendWinner = Form.useWatch(['test_settings', 'auto_send_winner'], form)

  // When auto-send winner is enabled, ensure tracking is enabled
  useEffect(() => {
    if (autoSendWinner) {
      form.setFieldValue('tracking_enabled', true)
    }
  }, [autoSendWinner, form])

  if (!autoSendWinner) return null

  return (
    <Row gutter={24}>
      <Col span={12}>
        <Form.Item
          name={['test_settings', 'auto_send_winner_metric']}
          label="Winning metric"
          rules={[{ required: true }]}
        >
          <Select
            options={[
              { value: 'open_rate', label: 'Open Rate' },
              { value: 'click_rate', label: 'Click Rate' }
            ]}
          />
        </Form.Item>
      </Col>
      <Col span={12}>
        <Form.Item
          name={['test_settings', 'test_duration_hours']}
          label="Test duration (hours)"
          rules={[{ required: true }]}
        >
          <InputNumber min={1} />
        </Form.Item>
      </Col>
    </Row>
  )
}

// Custom component to handle tracking enabled with disabled state
const TrackingEnabledField = ({ form }: { form: any }) => {
  const autoSendWinner = Form.useWatch(['test_settings', 'auto_send_winner'], form)

  return (
    <Form.Item
      name="tracking_enabled"
      label="Enable tracking"
      valuePropName="checked"
      tooltip="Must be enabled when using auto-send winner feature"
    >
      <Switch disabled={autoSendWinner} />
    </Form.Item>
  )
}

interface UpsertBroadcastDrawerProps {
  workspace: Workspace
  broadcast?: Broadcast
  buttonProps?: any
  buttonContent?: React.ReactNode
  onClose?: () => void
}

export function UpsertBroadcastDrawer({
  workspace,
  broadcast,
  buttonProps = {},
  buttonContent,
  onClose
}: UpsertBroadcastDrawerProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()
  const [tab, setTab] = useState<string>('settings')
  const [loading, setLoading] = useState(false)
  const { message } = App.useApp()
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/campaigns' })

  // Watch campaign name changes using Form.useWatch
  const campaignName = Form.useWatch('name', form)

  // Update utm_campaign when campaign name changes
  useEffect(() => {
    if (campaignName && isOpen) {
      // Convert to snake_case: lowercase, replace spaces and special chars with underscore
      const snakeCaseName = campaignName
        .toLowerCase()
        .replace(/[^\w\s]/g, '_') // Replace special characters with underscore
        .replace(/\s+/g, '_') // Replace spaces with underscore
        .replace(/_+/g, '_') // Replace multiple underscores with a single one

      // Set the utm_campaign value
      form.setFieldValue(['utm_parameters', 'campaign'], snakeCaseName)
    }
  }, [campaignName, form, isOpen])

  // Fetch templates for the dropdown selection
  const { data: templatesData } = useQuery({
    queryKey: ['templates', workspace.id],
    queryFn: () => {
      return templatesApi.list({ workspace_id: workspace.id })
    },
    enabled: isOpen
  })

  const upsertBroadcastMutation = useMutation({
    mutationFn: (values: CreateBroadcastRequest | UpdateBroadcastRequest) => {
      if (broadcast) {
        return broadcastApi.update(values as UpdateBroadcastRequest)
      } else {
        return broadcastApi.create(values as CreateBroadcastRequest)
      }
    },
    onSuccess: () => {
      message.success(`Campaign ${broadcast ? 'updated' : 'created'} successfully`)
      handleClose()
      queryClient.invalidateQueries({ queryKey: ['broadcasts', workspace.id] })
      setLoading(false)
    },
    onError: (error) => {
      message.error(`Failed to ${broadcast ? 'update' : 'create'} campaign: ${error.message}`)
      setLoading(false)
    }
  })

  const showDrawer = () => {
    if (broadcast) {
      form.setFieldsValue({
        id: broadcast.id,
        name: broadcast.name,
        audience: broadcast.audience,
        schedule: broadcast.schedule,
        test_settings: broadcast.test_settings,
        goal_id: broadcast.goal_id || undefined,
        tracking_enabled: broadcast.tracking_enabled,
        utm_parameters: broadcast.utm_parameters || undefined,
        metadata: broadcast.metadata || undefined
      })
    } else {
      // Set default values for a new broadcast
      form.setFieldsValue({
        name: '',
        audience: {
          lists: [],
          segments: [],
          exclude_unsubscribed: true,
          skip_duplicate_emails: true
        },
        schedule: {
          send_immediately: true,
          use_recipient_timezone: false
        },
        test_settings: {
          enabled: false,
          sample_percentage: 50,
          auto_send_winner: false,
          variations: [
            {
              id: 'default',
              name: 'Default',
              template_id: '',
              template_version: 1,
              subject: '',
              from_name: '',
              from_email: ''
            }
          ]
        },
        tracking_enabled: true,
        utm_parameters: {
          medium: 'email'
        }
      })
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

  const goToSettings = () => {
    setTab('settings')
  }

  const goToTemplate = () => {
    setTab('template')
  }

  const goToSchedule = () => {
    setTab('schedule')
  }

  const renderTabExtra = () => {
    return (
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
            <>
              <Button type="primary" ghost onClick={goToSettings}>
                Previous
              </Button>
              <Button type="primary" onClick={goToSchedule}>
                Next
              </Button>
            </>
          )}

          {tab === 'schedule' && (
            <>
              <Button type="primary" ghost onClick={goToTemplate}>
                Previous
              </Button>
              <Button
                loading={loading || upsertBroadcastMutation.isPending}
                onClick={() => {
                  form.submit()
                }}
                type="primary"
              >
                Save
              </Button>
            </>
          )}
        </Space>
      </div>
    )
  }

  return (
    <>
      <Button type="primary" onClick={showDrawer} {...buttonProps}>
        {buttonContent || (broadcast ? 'Edit Campaign' : 'Create Campaign')}
      </Button>
      {isOpen && (
        <Drawer
          title={<>{broadcast ? 'Edit campaign' : 'Create a campaign'}</>}
          closable={true}
          keyboard={false}
          maskClosable={false}
          width={'95%'}
          open={isOpen}
          onClose={handleClose}
          className="drawer-no-transition drawer-body-no-padding"
          extra={renderTabExtra()}
        >
          <Form
            form={form}
            layout="vertical"
            onFinish={(values) => {
              setLoading(true)

              // Ensure workspace_id is included
              const payload = {
                ...values,
                workspace_id: workspace.id
              }

              // Add ID for updates
              if (broadcast) {
                payload.id = broadcast.id
              }

              upsertBroadcastMutation.mutate(payload)
            }}
            onFinishFailed={(info) => {
              if (info.errorFields) {
                // Navigate to the tab with errors
                const errorFieldPaths = info.errorFields.map((field) => field.name.join('.'))

                // Determine which tab contains errors
                if (
                  errorFieldPaths.some(
                    (path) => path.startsWith('name') || path.startsWith('audience')
                  )
                ) {
                  setTab('settings')
                } else if (errorFieldPaths.some((path) => path.startsWith('test_settings'))) {
                  setTab('template')
                } else if (errorFieldPaths.some((path) => path.startsWith('schedule'))) {
                  setTab('schedule')
                }
              }
              setLoading(false)
            }}
          >
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
                    label: '1. Settings & Audience'
                  },
                  {
                    key: 'template',
                    label: '2. Template'
                  },
                  {
                    key: 'schedule',
                    label: '3. Schedule'
                  }
                ]}
              />
            </div>

            <div className="relative">
              {/* Settings & Audience Tab */}
              <div style={{ display: tab === 'settings' ? 'block' : 'none' }}>
                <div className="p-8">
                  <div className="text-lg mb-6 font-bold">Campaign Settings</div>
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        name="name"
                        label="Campaign name"
                        rules={[{ required: true, message: 'Please enter a campaign name' }]}
                      >
                        <Input placeholder="E.g. Weekly Newsletter - May 2023" />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <TrackingEnabledField form={form} />
                    </Col>
                  </Row>

                  <div className="text-lg mt-8 mb-6 font-bold">Audience Selection</div>

                  <Form.Item
                    noStyle
                    dependencies={[['audience', 'lists']]}
                    rules={[
                      {
                        validator: async (_, form) => {
                          const lists = form.getFieldValue(['audience', 'lists']) || []

                          if (lists.length === 0) {
                            return Promise.reject(
                              new Error('Please select lists for your audience')
                            )
                          }

                          return Promise.resolve()
                        }
                      }
                    ]}
                  >
                    {() => null}
                  </Form.Item>

                  <Form.Item
                    name={['audience', 'lists']}
                    label="Lists"
                    extra="Select the contact lists to include in this campaign"
                  >
                    <Select
                      mode="multiple"
                      placeholder="Select lists"
                      options={[
                        // These would be fetched from an API in a real implementation
                        { value: 'list-1', label: 'Marketing Contacts' },
                        { value: 'list-2', label: 'Newsletter Subscribers' },
                        { value: 'list-3', label: 'New Customers' }
                      ]}
                    />
                  </Form.Item>

                  <div className="text-lg mt-8 mb-4 font-bold">Advanced Options</div>
                  <Row gutter={24}>
                    <Col span={12}>
                      <Form.Item
                        name={['audience', 'exclude_unsubscribed']}
                        label="Exclude unsubscribed recipients"
                        valuePropName="checked"
                        initialValue={true}
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item
                        name={['audience', 'skip_duplicate_emails']}
                        label="Skip duplicate emails"
                        valuePropName="checked"
                        initialValue={true}
                      >
                        <Switch />
                      </Form.Item>
                    </Col>
                  </Row>

                  <Form.Item
                    name={['audience', 'rate_limit_per_minute']}
                    label="Rate limit (emails per minute)"
                  >
                    <InputNumber min={1} />
                  </Form.Item>

                  <div className="text-lg mt-8 mb-4 font-bold">URL Tracking Parameters</div>
                  <Row gutter={24}>
                    <Col span={8}>
                      <Form.Item name={['utm_parameters', 'source']} label="utm_source">
                        <Input placeholder="Your website or company name" />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        name={['utm_parameters', 'medium']}
                        label="utm_medium"
                        initialValue="email"
                      >
                        <Input placeholder="email" />
                      </Form.Item>
                    </Col>
                    <Col span={8}>
                      <Form.Item
                        name={['utm_parameters', 'campaign']}
                        label="utm_campaign"
                        tooltip="Automatically generated from campaign name"
                      >
                        <Input placeholder="Generated from campaign name" />
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              </div>

              {/* Template Tab */}
              <div style={{ display: tab === 'template' ? 'block' : 'none' }}>
                <div className="p-8">
                  <Form.Item
                    name={['test_settings', 'enabled']}
                    label="Enable A/B Testing"
                    valuePropName="checked"
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    noStyle
                    shouldUpdate={(prevValues, currentValues) => {
                      return (
                        prevValues.test_settings?.enabled !== currentValues.test_settings?.enabled
                      )
                    }}
                  >
                    {({ getFieldValue }) => {
                      const testEnabled = getFieldValue(['test_settings', 'enabled'])

                      if (testEnabled) {
                        return (
                          <>
                            <Row gutter={24}>
                              <Col span={12}>
                                <Form.Item
                                  name={['test_settings', 'sample_percentage']}
                                  label="Test sample size (%)"
                                  rules={[{ required: true }]}
                                >
                                  <InputNumber min={1} max={100} />
                                </Form.Item>
                              </Col>
                              <Col span={12}>
                                <Form.Item
                                  name={['test_settings', 'auto_send_winner']}
                                  label="Automatically send winner"
                                  valuePropName="checked"
                                  tooltip="Requires tracking to be enabled"
                                >
                                  <Switch />
                                </Form.Item>
                              </Col>
                            </Row>

                            <ABTestingConfig form={form} />

                            {/* Variations management will be added here */}
                            <div className="text-lg mt-4 mb-4 font-bold">Variations</div>

                            <Form.List name={['test_settings', 'variations']}>
                              {(fields, { add, remove }) => (
                                <>
                                  {fields.map((field) => (
                                    <div key={field.key} className="border p-4 mb-4 rounded">
                                      <Row gutter={24}>
                                        <Col span={11}>
                                          <Form.Item
                                            key={`name-${field.key}`}
                                            name={[field.name, 'name']}
                                            label="Variation name"
                                            rules={[{ required: true }]}
                                          >
                                            <Input placeholder="E.g. Variation A" />
                                          </Form.Item>
                                        </Col>
                                        <Col span={11}>
                                          <Form.Item
                                            key={`template-${field.key}`}
                                            name={[field.name, 'template_id']}
                                            label="Template"
                                            rules={[{ required: true }]}
                                          >
                                            <TemplateSelectorInput
                                              workspaceId={workspace.id}
                                              placeholder="Select template"
                                            />
                                          </Form.Item>
                                        </Col>
                                        {fields.length > 1 && (
                                          <Col span={2} className="flex items-end justify-end pb-2">
                                            <Form.Item label=" ">
                                              <Popconfirm
                                                title="Remove variation"
                                                description="Are you sure you want to remove this variation?"
                                                onConfirm={() => remove(field.name)}
                                                okText="Yes"
                                                cancelText="No"
                                              >
                                                <Button
                                                  type="text"
                                                  danger
                                                  icon={<DeleteOutlined />}
                                                />
                                              </Popconfirm>
                                            </Form.Item>
                                          </Col>
                                        )}
                                      </Row>
                                    </div>
                                  ))}

                                  {fields.length < 5 && (
                                    <Button
                                      type="dashed"
                                      onClick={() =>
                                        add({
                                          id: `variation-${fields.length + 1}`,
                                          name: `Variation ${String.fromCharCode(65 + fields.length)}`,
                                          template_id: '',
                                          template_version: 1
                                        })
                                      }
                                      block
                                    >
                                      + Add variation
                                    </Button>
                                  )}
                                </>
                              )}
                            </Form.List>
                          </>
                        )
                      }

                      // If A/B testing is disabled, show single template config
                      return (
                        <div>
                          <Row gutter={24}>
                            <Col span={12}>
                              <Form.Item
                                name={['test_settings', 'variations', 0, 'template_id']}
                                label="Template"
                                rules={[{ required: true }]}
                              >
                                <TemplateSelectorInput
                                  workspaceId={workspace.id}
                                  placeholder="Select template"
                                />
                              </Form.Item>
                            </Col>
                          </Row>
                        </div>
                      )
                    }}
                  </Form.Item>
                </div>
              </div>

              {/* Schedule Tab */}
              <div style={{ display: tab === 'schedule' ? 'block' : 'none' }}>
                <div className="p-8">
                  <Form.Item
                    name={['schedule', 'send_immediately']}
                    valuePropName="checked"
                    label="Send immediately after saving"
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    noStyle
                    shouldUpdate={(prevValues, currentValues) => {
                      return (
                        prevValues.schedule?.send_immediately !==
                        currentValues.schedule?.send_immediately
                      )
                    }}
                  >
                    {({ getFieldValue }) => {
                      const sendImmediately = getFieldValue(['schedule', 'send_immediately'])

                      if (!sendImmediately) {
                        return (
                          <>
                            <Form.Item
                              name={['schedule', 'scheduled_time']}
                              label="Schedule date and time"
                              rules={[{ required: true, message: 'Please select a date and time' }]}
                            >
                              <DatePicker
                                showTime
                                format="YYYY-MM-DD HH:mm"
                                disabledDate={(current) => {
                                  // Can't select days before today
                                  return current && current < dayjs().startOf('day')
                                }}
                              />
                            </Form.Item>

                            <Form.Item
                              name={['schedule', 'use_recipient_timezone']}
                              valuePropName="checked"
                              label="Send according to recipient timezone"
                            >
                              <Switch />
                            </Form.Item>

                            <Form.Item
                              noStyle
                              shouldUpdate={(prevValues, currentValues) => {
                                return (
                                  prevValues.schedule?.use_recipient_timezone !==
                                  currentValues.schedule?.use_recipient_timezone
                                )
                              }}
                            >
                              {({ getFieldValue }) => {
                                const useRecipientTimezone = getFieldValue([
                                  'schedule',
                                  'use_recipient_timezone'
                                ])

                                if (useRecipientTimezone) {
                                  return (
                                    <Row gutter={24}>
                                      <Col span={12}>
                                        <Form.Item
                                          name={['schedule', 'time_window_start']}
                                          label="Delivery window start"
                                          rules={[{ required: true }]}
                                        >
                                          <Select
                                            options={Array.from({ length: 24 }, (_, i) => ({
                                              value: `${i}:00`,
                                              label: `${i}:00`
                                            }))}
                                          />
                                        </Form.Item>
                                      </Col>
                                      <Col span={12}>
                                        <Form.Item
                                          name={['schedule', 'time_window_end']}
                                          label="Delivery window end"
                                          rules={[{ required: true }]}
                                        >
                                          <Select
                                            options={Array.from({ length: 24 }, (_, i) => ({
                                              value: `${i}:00`,
                                              label: `${i}:00`
                                            }))}
                                          />
                                        </Form.Item>
                                      </Col>
                                    </Row>
                                  )
                                }

                                return null
                              }}
                            </Form.Item>
                          </>
                        )
                      }

                      return null
                    }}
                  </Form.Item>
                </div>
              </div>
            </div>
          </Form>
        </Drawer>
      )}
    </>
  )
}
