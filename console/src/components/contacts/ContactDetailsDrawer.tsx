import React from 'react'
import {
  Drawer,
  Space,
  Tag,
  Typography,
  Table,
  Badge,
  Spin,
  Empty,
  Tooltip,
  Button,
  Modal,
  Select,
  Form,
  message
} from 'antd'
import { Contact } from '../../services/api/contacts'
import { List } from '../../services/api/types'
import dayjs from '../../lib/dayjs'
import { ContactUpsertDrawer } from './ContactUpsertDrawer'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faCalendar,
  faShoppingCart,
  faMoneyBillWave,
  faPlus
} from '@fortawesome/free-solid-svg-icons'
import { faPenToSquare } from '@fortawesome/free-regular-svg-icons'
import { useQuery, useQueryClient, useMutation } from '@tanstack/react-query'
import { listMessages, MessageStatus } from '../../services/api/messages_history'
import { contactsApi } from '../../services/api/contacts'
import {
  contactListApi,
  UpdateContactListStatusRequest,
  AddContactToListRequest
} from '../../services/api/contact_list'

const { Title, Text } = Typography

interface ContactDetailsDrawerProps {
  workspaceId: string
  contact?: Contact
  visible: boolean
  onClose: () => void
  lists?: List[]
  onContactUpdated?: (updatedContact: Contact) => void
}

// Add this type definition for the lists with name
interface ContactListWithName {
  list_id: string
  status: string
  name: string
}

export function ContactDetailsDrawer({
  workspaceId,
  contact,
  visible,
  onClose,
  lists = [],
  onContactUpdated
}: ContactDetailsDrawerProps) {
  if (!contact) return null

  const queryClient = useQueryClient()
  const [statusModalVisible, setStatusModalVisible] = React.useState(false)
  const [subscribeModalVisible, setSubscribeModalVisible] = React.useState(false)
  const [selectedList, setSelectedList] = React.useState<ContactListWithName | null>(null)
  const [statusForm] = Form.useForm()
  const [subscribeForm] = Form.useForm()

  // Load message history for this contact
  const { data: messageHistory, isLoading: loadingMessages } = useQuery({
    queryKey: ['message_history', workspaceId, contact.email],
    queryFn: () =>
      listMessages(workspaceId, {
        contact_id: contact.email,
        limit: 50
      }),
    enabled: visible && !!contact
  })

  // Fetch the single contact to ensure we have the latest data
  const { data: refreshedContact, isLoading: isLoadingContact } = useQuery({
    queryKey: ['contact_details', workspaceId, contact.email],
    queryFn: async () => {
      const response = await contactsApi.list({
        workspace_id: workspaceId,
        email: contact.email,
        with_contact_lists: true,
        limit: 1
      })
      return response.contacts[0]
    },
    enabled: visible && !!contact,
    refetchOnWindowFocus: true
  })

  // Update parent component when contact is refreshed
  React.useEffect(() => {
    if (refreshedContact && onContactUpdated) {
      onContactUpdated(refreshedContact)
    }
  }, [refreshedContact, onContactUpdated])

  // Mutation for updating subscription status
  const updateStatusMutation = useMutation({
    mutationFn: (params: UpdateContactListStatusRequest) => contactListApi.updateStatus(params),
    onSuccess: () => {
      message.success('Subscription status updated successfully')
      queryClient.invalidateQueries({ queryKey: ['contact_details', workspaceId, contact.email] })
      setStatusModalVisible(false)
      statusForm.resetFields()
    },
    onError: (error) => {
      message.error(`Failed to update status: ${error}`)
    }
  })

  // Mutation for adding contact to a list
  const addToListMutation = useMutation({
    mutationFn: (params: AddContactToListRequest) => contactListApi.addContact(params),
    onSuccess: () => {
      message.success('Contact added to list successfully')
      queryClient.invalidateQueries({ queryKey: ['contact_details', workspaceId, contact.email] })
      setSubscribeModalVisible(false)
      subscribeForm.resetFields()
    },
    onError: (error) => {
      message.error(`Failed to add to list: ${error}`)
    }
  })

  const handleContactUpdated = () => {
    // Invalidate both the contact details and the contacts list queries
    queryClient.invalidateQueries({ queryKey: ['contact_details', workspaceId, contact.email] })
    queryClient.invalidateQueries({ queryKey: ['contacts', workspaceId] })
  }

  // Use the refreshed contact data if available, otherwise fall back to the original contact
  const displayContact = refreshedContact || contact

  // Find list names based on list IDs
  const getListName = (listId: string): string => {
    const list = lists.find((list) => list.id === listId)
    return list ? list.name : listId
  }

  // Handle opening the status change modal
  const openStatusModal = (list: ContactListWithName) => {
    setSelectedList(list)
    statusForm.setFieldsValue({
      status: list.status
    })
    setStatusModalVisible(true)
  }

  // Handle status change submission
  const handleStatusChange = (values: { status: string }) => {
    if (!selectedList) return

    updateStatusMutation.mutate({
      workspace_id: workspaceId,
      email: displayContact.email,
      list_id: selectedList.list_id,
      status: values.status
    })
  }

  // Handle opening the subscribe to list modal
  const openSubscribeModal = () => {
    subscribeForm.resetFields()
    setSubscribeModalVisible(true)
  }

  // Handle subscribe to list submission
  const handleSubscribe = (values: { list_id: string; status: string }) => {
    addToListMutation.mutate({
      workspace_id: workspaceId,
      email: displayContact.email,
      list_id: values.list_id,
      status: values.status as 'active' | 'pending'
    })
  }

  // Create name from first and last name
  const fullName =
    [displayContact.first_name, displayContact.last_name].filter(Boolean).join(' ') || 'Unknown'

  const formatValue = (value: any) => {
    if (value === null || value === undefined) return '-'
    if (typeof value === 'object') return JSON.stringify(value, null, 2)
    return value
  }

  // Format date using dayjs
  const formatDate = (dateString: string | undefined): string => {
    if (!dateString) return '-'
    return dayjs(dateString).format('lll')
  }

  // Format currency value
  const formatCurrency = (value: number | undefined): string => {
    if (value === undefined || value === null) return '$0.00'
    return `$${value.toFixed(2)}`
  }

  // Get color for list status
  const getStatusColor = (status: string): string => {
    const statusColors: Record<string, string> = {
      active: 'green',
      subscribed: 'green',
      pending: 'orange',
      unsubscribed: 'red',
      bounced: 'volcano',
      complained: 'magenta',
      blacklisted: 'black'
    }
    return statusColors[status.toLowerCase()] || 'blue'
  }

  // Status badge for message history
  const getStatusBadge = (status: MessageStatus) => {
    const statusConfig: Record<MessageStatus, { color: string; text: string }> = {
      sent: { color: 'blue', text: 'Sent' },
      delivered: { color: 'green', text: 'Delivered' },
      failed: { color: 'red', text: 'Failed' },
      opened: { color: 'cyan', text: 'Opened' },
      clicked: { color: 'geekblue', text: 'Clicked' },
      bounced: { color: 'volcano', text: 'Bounced' },
      complained: { color: 'magenta', text: 'Complained' },
      unsubscribed: { color: 'gold', text: 'Unsubscribed' }
    }

    const config = statusConfig[status] || { color: 'default', text: status }
    return <Badge status={config.color as any} text={config.text} />
  }

  // Table columns for message history
  const messageColumns = [
    {
      title: 'Channel',
      dataIndex: 'channel',
      key: 'channel',
      width: '15%'
    },
    {
      title: 'Template',
      dataIndex: 'template_id',
      key: 'template_id',
      width: '20%'
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      width: '15%',
      render: (status: MessageStatus) => getStatusBadge(status)
    },
    {
      title: 'Sent At',
      dataIndex: 'sent_at',
      key: 'sent_at',
      width: '20%',
      render: (date: string) => formatDate(date)
    },
    {
      title: 'Last Update',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: '20%',
      render: (date: string) => formatDate(date)
    }
  ]

  // Field display definitions without icons
  const contactFields = [
    { key: 'first_name', label: 'First Name', value: displayContact.first_name },
    { key: 'last_name', label: 'Last Name', value: displayContact.last_name },
    { key: 'email', label: 'Email', value: displayContact.email },
    { key: 'phone', label: 'Phone', value: displayContact.phone },
    {
      key: 'address',
      label: 'Address',
      value: [
        displayContact.address_line_1,
        displayContact.address_line_2,
        [displayContact.state, displayContact.postcode, displayContact.country]
          .filter(Boolean)
          .join(', ')
      ]
        .filter(Boolean)
        .join(', '),
      show: !!(
        displayContact.address_line_1 ||
        displayContact.address_line_2 ||
        displayContact.country ||
        displayContact.state ||
        displayContact.postcode
      )
    },
    { key: 'job_title', label: 'Job Title', value: displayContact.job_title },
    { key: 'timezone', label: 'Timezone', value: displayContact.timezone },
    { key: 'language', label: 'Language', value: displayContact.language },
    { key: 'external_id', label: 'External ID', value: displayContact.external_id },
    {
      key: 'lifetime_value',
      label: 'Lifetime Value',
      value: displayContact.lifetime_value
    },
    {
      key: 'orders_count',
      label: 'Orders Count',
      value: displayContact.orders_count
    },
    {
      key: 'last_order_at',
      label: 'Last Order At',
      value: formatDate(displayContact.last_order_at)
    },
    {
      key: 'created_at',
      label: 'Created At',
      value: formatDate(displayContact.created_at)
    },
    {
      key: 'updated_at',
      label: 'Updated At',
      value: formatDate(displayContact.updated_at)
    },
    // Custom string fields
    {
      key: 'custom_string_1',
      label: 'Custom String 1',
      value: displayContact.custom_string_1
    },
    {
      key: 'custom_string_2',
      label: 'Custom String 2',
      value: displayContact.custom_string_2
    },
    {
      key: 'custom_string_3',
      label: 'Custom String 3',
      value: displayContact.custom_string_3
    },
    {
      key: 'custom_string_4',
      label: 'Custom String 4',
      value: displayContact.custom_string_4
    },
    {
      key: 'custom_string_5',
      label: 'Custom String 5',
      value: displayContact.custom_string_5
    },
    // Custom number fields
    {
      key: 'custom_number_1',
      label: 'Custom Number 1',
      value: displayContact.custom_number_1
    },
    {
      key: 'custom_number_2',
      label: 'Custom Number 2',
      value: displayContact.custom_number_2
    },
    {
      key: 'custom_number_3',
      label: 'Custom Number 3',
      value: displayContact.custom_number_3
    },
    {
      key: 'custom_number_4',
      label: 'Custom Number 4',
      value: displayContact.custom_number_4
    },
    {
      key: 'custom_number_5',
      label: 'Custom Number 5',
      value: displayContact.custom_number_5
    },
    // Custom date fields
    {
      key: 'custom_datetime_1',
      label: 'Custom Date 1',
      value: formatDate(displayContact.custom_datetime_1)
    },
    {
      key: 'custom_datetime_2',
      label: 'Custom Date 2',
      value: formatDate(displayContact.custom_datetime_2)
    },
    {
      key: 'custom_datetime_3',
      label: 'Custom Date 3',
      value: formatDate(displayContact.custom_datetime_3)
    },
    {
      key: 'custom_datetime_4',
      label: 'Custom Date 4',
      value: formatDate(displayContact.custom_datetime_4)
    },
    {
      key: 'custom_datetime_5',
      label: 'Custom Date 5',
      value: formatDate(displayContact.custom_datetime_5)
    }
  ]

  // Prepare contact lists with enhanced information
  const contactListsWithNames = displayContact.contact_lists.map((list) => ({
    ...list,
    name: getListName(list.list_id)
  }))

  // Get lists that the contact is not subscribed to
  const availableLists = lists.filter(
    (list) => !displayContact.contact_lists.some((cl) => cl.list_id === list.id)
  )

  // Status options for dropdown
  const statusOptions = [
    { label: 'Active', value: 'active' },
    { label: 'Pending', value: 'pending' },
    { label: 'Unsubscribed', value: 'unsubscribed' },
    { label: 'Blacklisted', value: 'blacklisted' }
  ]

  return (
    <Drawer
      title="Contact Details"
      width={1000}
      placement="right"
      className="drawer-body-no-padding"
      onClose={onClose}
      open={visible}
      extra={
        <ContactUpsertDrawer
          workspaceId={workspaceId}
          contact={displayContact}
          onSuccess={handleContactUpdated}
          buttonProps={{
            icon: <FontAwesomeIcon icon={faPenToSquare} />,
            type: 'text',
            buttonContent: 'Update'
          }}
        />
      }
    >
      <div className="flex h-full">
        {/* Left column - Contact Details (1/4 width) */}
        <div className="w-1/4 bg-gray-50 overflow-y-auto h-full">
          {/* Contact info at the top */}
          <div className="p-6 pb-4 border-b border-gray-200 flex flex-col items-center text-center">
            <Title level={4} style={{ margin: 0, marginBottom: '4px' }}>
              {fullName}
            </Title>
            <Text type="secondary">{displayContact.email}</Text>
          </div>

          <div className="p-6">
            <div className="contact-details">
              {isLoadingContact && (
                <div className="mb-4 p-2 bg-blue-50 text-blue-600 rounded text-center">
                  <Spin size="small" className="mr-2" />
                  <span>Refreshing contact data...</span>
                </div>
              )}

              {/* Flat list of all fields without icons */}
              <Space direction="vertical" size="middle" style={{ width: '100%' }}>
                {contactFields
                  .filter(
                    (field) =>
                      field.value !== undefined &&
                      field.value !== null &&
                      field.value !== '-' &&
                      (field.show === undefined || field.show) &&
                      // Skip email as it's already shown at the top
                      field.key !== 'email' &&
                      // Skip name fields as they're already shown at the top
                      field.key !== 'first_name' &&
                      field.key !== 'last_name'
                  )
                  .map((field) => (
                    <div key={field.key}>
                      <Tooltip title={`Field ID: ${field.key}`}>
                        <Text strong style={{ cursor: 'help' }}>
                          {field.label}:
                        </Text>
                      </Tooltip>
                      <div>
                        <Text>{formatValue(field.value)}</Text>
                      </div>
                    </div>
                  ))}
              </Space>
            </div>
          </div>
        </div>

        {/* Right column - Message History (3/4 width) */}
        <div className="w-3/4 p-6 overflow-y-auto h-full">
          {/* E-commerce Stats (3-column grid) */}
          <div className="grid grid-cols-3 gap-4 mb-6">
            {/* Lifetime Value */}
            <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
              <div className="text-sm text-gray-500 mb-2">
                <Tooltip title="Field ID: lifetime_value">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faMoneyBillWave} className="mr-2" />
                    Lifetime Value
                  </span>
                </Tooltip>
              </div>
              <div className="text-2xl font-semibold">
                {formatCurrency(displayContact.lifetime_value)}
              </div>
            </div>

            {/* Orders Count */}
            <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
              <div className="text-sm text-gray-500 mb-2">
                <Tooltip title="Field ID: orders_count">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faShoppingCart} className="mr-2" />
                    Orders Count
                  </span>
                </Tooltip>
              </div>
              <div className="text-2xl font-semibold">{displayContact.orders_count || 0}</div>
            </div>

            {/* Last Order */}
            <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
              <div className="text-sm text-gray-500 mb-2">
                <Tooltip title="Field ID: last_order_at">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faCalendar} className="mr-2" />
                    Last Order
                  </span>
                </Tooltip>
              </div>
              <div className="text-lg font-semibold">
                {displayContact.last_order_at
                  ? dayjs(displayContact.last_order_at).format('ll')
                  : 'Never'}
              </div>
            </div>
          </div>

          {/* List subscriptions with action buttons */}
          <div className="flex justify-between items-center mb-3">
            <Title level={5} style={{ margin: 0 }}>
              List Subscriptions
            </Title>
            <Button
              type="primary"
              ghost
              size="small"
              icon={<FontAwesomeIcon icon={faPlus} />}
              onClick={openSubscribeModal}
              disabled={availableLists.length === 0}
            >
              Subscribe to List
            </Button>
          </div>

          <Table
            dataSource={contactListsWithNames}
            rowKey={(record) => `${record.list_id}_${record.status}`}
            pagination={false}
            size="small"
            columns={[
              {
                title: 'Subscription list',
                dataIndex: 'name',
                key: 'name',
                width: '50%',
                render: (name: string, record: any) => (
                  <Tooltip title={`List ID: ${record.list_id}`}>
                    <span style={{ cursor: 'help' }}>{name}</span>
                  </Tooltip>
                )
              },
              {
                title: 'Status',
                dataIndex: 'status',
                key: 'status',
                width: '30%',
                render: (status: string) => <Tag color={getStatusColor(status)}>{status}</Tag>
              },
              {
                title: '',
                key: 'actions',
                width: '20%',
                render: (_: any, record: ContactListWithName) => (
                  <Button
                    size="small"
                    onClick={() => openStatusModal(record)}
                    loading={
                      updateStatusMutation.isPending && selectedList?.list_id === record.list_id
                    }
                  >
                    Change Status
                  </Button>
                )
              }
            ]}
          />

          <div className="mt-6">
            <div className="section-header">
              <Space>
                <Title level={5} style={{ margin: 0 }}>
                  Message History
                </Title>
              </Space>
            </div>

            {loadingMessages ? (
              <div className="loading-container" style={{ padding: '40px 0', textAlign: 'center' }}>
                <Spin size="large" />
                <div style={{ marginTop: 16 }}>Loading message history...</div>
              </div>
            ) : messageHistory && messageHistory.messages.length > 0 ? (
              <Table
                dataSource={messageHistory.messages}
                columns={messageColumns}
                rowKey="id"
                pagination={false}
                size="small"
                scroll={{ y: 600 }}
              />
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No messages found for this contact"
                style={{ margin: '40px 0' }}
              />
            )}
          </div>
        </div>
      </div>

      {/* Change Status Modal */}
      <Modal
        title={`Change Status for ${selectedList?.name || 'List'}`}
        open={statusModalVisible}
        onCancel={() => setStatusModalVisible(false)}
        footer={null}
      >
        <Form form={statusForm} layout="vertical" onFinish={handleStatusChange}>
          <Form.Item
            name="status"
            label="Subscription Status"
            rules={[{ required: true, message: 'Please select a status' }]}
          >
            <Select options={statusOptions} />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={updateStatusMutation.isPending}>
                Update Status
              </Button>
              <Button onClick={() => setStatusModalVisible(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>

      {/* Subscribe to List Modal */}
      <Modal
        title="Subscribe to List"
        open={subscribeModalVisible}
        onCancel={() => setSubscribeModalVisible(false)}
        footer={null}
      >
        <Form form={subscribeForm} layout="vertical" onFinish={handleSubscribe}>
          <Form.Item
            name="list_id"
            label="Select List"
            rules={[{ required: true, message: 'Please select a list' }]}
          >
            <Select
              options={availableLists.map((list) => ({
                label: list.name,
                value: list.id
              }))}
              placeholder="Select a list"
            />
          </Form.Item>
          <Form.Item
            name="status"
            label="Subscription Status"
            initialValue="active"
            rules={[{ required: true, message: 'Please select a status' }]}
          >
            <Select
              options={[
                { label: 'Active', value: 'active' },
                { label: 'Pending', value: 'pending' }
              ]}
            />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={addToListMutation.isPending}>
                Subscribe
              </Button>
              <Button onClick={() => setSubscribeModalVisible(false)}>Cancel</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Drawer>
  )
}
