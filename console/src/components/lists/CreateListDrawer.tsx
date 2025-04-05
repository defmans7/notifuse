import React from 'react'
import { Button, Drawer, Form, Input, Switch, message, Tooltip } from 'antd'
import { PlusOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { listsApi } from '../../services/api/list'
import type { CreateListRequest } from '../../services/api/types'

interface CreateListDrawerProps {
  workspaceId: string
  buttonProps?: {
    type?: 'primary' | 'default' | 'link'
    buttonContent?: React.ReactNode
    size?: 'large' | 'middle' | 'small'
  }
}

export function CreateListDrawer({
  workspaceId,
  buttonProps = {
    type: 'primary',
    buttonContent: 'Create List',
    size: 'middle'
  }
}: CreateListDrawerProps) {
  const [open, setOpen] = React.useState(false)
  const [form] = Form.useForm()
  const queryClient = useQueryClient()

  // Generate list ID from name (alphanumeric only)
  const generateListId = (name: string) => {
    if (!name) return ''
    // Remove spaces and non-alphanumeric characters
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]/g, '')
      .substring(0, 32)
  }

  // Update generated ID when name changes
  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const name = e.target.value
    const id = generateListId(name)
    form.setFieldsValue({ id })
  }

  const createListMutation = useMutation({
    mutationFn: (data: CreateListRequest) => {
      return listsApi.create(data)
    },
    onSuccess: () => {
      message.success('List created successfully')
      queryClient.invalidateQueries({ queryKey: ['lists', workspaceId] })
      setOpen(false)
      form.resetFields()
    },
    onError: (error) => {
      message.error(`Failed to create list: ${error}`)
    }
  })

  const showDrawer = () => {
    setOpen(true)
  }

  const onClose = () => {
    setOpen(false)
    form.resetFields()
  }

  const onFinish = (values: any) => {
    const request: CreateListRequest = {
      workspace_id: workspaceId,
      id: values.id,
      name: values.name,
      is_double_optin: values.is_double_optin || false,
      is_public: values.is_public || false,
      description: values.description
    }

    createListMutation.mutate(request)
  }

  return (
    <>
      <Button
        type={buttonProps.type || 'primary'}
        onClick={showDrawer}
        icon={<PlusOutlined />}
        size={buttonProps.size}
      >
        {buttonProps.buttonContent || 'Create List'}
      </Button>
      <Drawer
        title="Create New List"
        width={400}
        onClose={onClose}
        open={open}
        bodyStyle={{ paddingBottom: 80 }}
        extra={
          <Button
            type="primary"
            onClick={() => form.submit()}
            loading={createListMutation.isPending}
          >
            Create
          </Button>
        }
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            is_double_optin: false,
            is_public: false
          }}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[
              { required: true, message: 'Please enter a list name' },
              { max: 255, message: 'Name must be less than 255 characters' }
            ]}
          >
            <Input placeholder="Enter list name" onChange={handleNameChange} />
          </Form.Item>

          <Form.Item
            name="id"
            label="List ID"
            rules={[
              { required: true, message: 'Please enter a list ID' },
              { pattern: /^[a-zA-Z0-9]+$/, message: 'ID must be alphanumeric' },
              { max: 32, message: 'ID must be less than 32 characters' }
            ]}
          >
            <Input placeholder="Enter a unique alphanumeric ID" />
          </Form.Item>

          <Form.Item name="description" label="Description">
            <Input.TextArea rows={4} placeholder="Enter list description" />
          </Form.Item>

          <Form.Item
            name="is_public"
            label={
              <span>
                Public &nbsp;
                <Tooltip title="Public lists are visible in the subscription center for users to subscribe to">
                  <InfoCircleOutlined />
                </Tooltip>
              </span>
            }
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="is_double_optin"
            label={
              <span>
                Double Opt-in &nbsp;
                <Tooltip title="When enabled, subscribers must confirm their subscription via email before being added to the list">
                  <InfoCircleOutlined />
                </Tooltip>
              </span>
            }
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
        </Form>
      </Drawer>
    </>
  )
}
