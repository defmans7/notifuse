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
  message,
  Popover
} from 'antd'
import { Contact } from '../../services/api/contacts'
import { List } from '../../services/api/types'
import dayjs from '../../lib/dayjs'
import numbro from 'numbro'
import { ContactUpsertDrawer } from './ContactUpsertDrawer'
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import {
  faCalendar,
  faShoppingCart,
  faMoneyBillWave,
  faPlus,
  faEllipsis
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
  workspaceTimezone?: string
}

// Add this type definition for the lists with name
interface ContactListWithName {
  list_id: string
  status: string
  name: string
  created_at?: string
}

export function ContactDetailsDrawer({
  workspaceId,
  contact,
  visible,
  onClose,
  lists = [],
  onContactUpdated,
  workspaceTimezone = 'UTC'
}: ContactDetailsDrawerProps) {
  if (!contact) return null

  const queryClient = useQueryClient()
  const [statusModalVisible, setStatusModalVisible] = React.useState(false)
  const [subscribeModalVisible, setSubscribeModalVisible] = React.useState(false)
  const [selectedList, setSelectedList] = React.useState<ContactListWithName | null>(null)
  const [statusForm] = Form.useForm()
  const [subscribeForm] = Form.useForm()

  // Keep track of the currently displayed contact
  const [displayContact, setDisplayContact] = React.useState<Contact>(contact)

  // Update the display contact whenever the input contact changes
  React.useEffect(() => {
    if (contact) {
      setDisplayContact(contact)
    }
  }, [contact])

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

  // Update displayed contact and parent component when contact is refreshed
  React.useEffect(() => {
    if (refreshedContact) {
      setDisplayContact(refreshedContact)
      if (onContactUpdated) {
        onContactUpdated(refreshedContact)
      }
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

    // Format number values with numbro
    if (typeof value === 'number') {
      // For currency-like fields
      if (String(value).includes('.') && value > 0) {
        return numbro(value).format({
          thousandSeparated: true,
          mantissa: 2,
          trimMantissa: true
        })
      }
      // For integer values
      return numbro(value).format({
        thousandSeparated: true,
        mantissa: 0
      })
    }

    if (typeof value === 'object') return JSON.stringify(value, null, 2)
    return value
  }

  // Format JSON with truncation and popover for full view
  const formatJson = (jsonData: any): React.ReactNode => {
    if (!jsonData) return '-'

    try {
      // If it's already an object, stringify it
      const jsonStr = typeof jsonData === 'string' ? jsonData : JSON.stringify(jsonData)
      const obj = typeof jsonData === 'string' ? JSON.parse(jsonData) : jsonData

      // Pretty format for popover
      const prettyJson = JSON.stringify(obj, null, 2)

      // Truncate for display
      const displayText = jsonStr.length > 100 ? jsonStr.substring(0, 100) + '...' : jsonStr

      const popoverContent = (
        <div
          className="p-2 bg-gray-50 rounded border border-gray-200 max-h-96 overflow-auto"
          style={{ maxWidth: '500px' }}
        >
          <pre className="text-xs m-0 whitespace-pre-wrap break-all">{prettyJson}</pre>
        </div>
      )

      return (
        <Popover
          content={popoverContent}
          title="JSON Data"
          placement="rightTop"
          trigger="click"
          styles={{
            root: {
              maxWidth: '400px'
            }
          }}
        >
          <div className="text-xs bg-gray-50 p-2 rounded border border-gray-200 cursor-pointer hover:bg-gray-100">
            <code className="truncate block">{displayText}</code>
            <div className="text-right mt-1 text-blue-500">
              <small>
                <FontAwesomeIcon icon={faEllipsis} className="mr-1" />
                Click to view full JSON
              </small>
            </div>
          </div>
        </Popover>
      )
    } catch (e) {
      return <Text type="danger">Invalid JSON</Text>
    }
  }

  // Format date using dayjs
  const formatDate = (dateString: string | undefined): string => {
    if (!dateString) return '-'
    return `${dayjs(dateString).format('lll')} in ${workspaceTimezone}`
  }

  // Format currency value using numbro
  const formatCurrency = (value: number | undefined): string => {
    if (value === undefined || value === null) return '$0.00'
    return numbro(value).formatCurrency({
      mantissa: 2,
      currencySymbol: '$',
      thousandSeparated: true,
      trimMantissa: true,
      spaceSeparatedCurrency: false
    })
  }

  // Format number with thousand separators
  const formatNumber = (value: number | undefined): string => {
    if (value === undefined || value === null) return '0'
    return numbro(value).format({
      thousandSeparated: true,
      mantissa: 0,
      trimMantissa: true,
      average: false
    })
  }

  // Format average number (with K, M, B, etc. for large numbers)
  const formatAverage = (value: number | undefined): string => {
    if (value === undefined || value === null) return '0'
    return numbro(value).format({
      average: true,
      mantissa: 1,
      spaceSeparated: true,
      trimMantissa: true
    })
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

  // Add a separate section for JSON fields
  const jsonFields = [
    {
      key: 'custom_json_1',
      label: 'Custom JSON 1',
      value: displayContact.custom_json_1,
      show: !!displayContact.custom_json_1
    },
    {
      key: 'custom_json_2',
      label: 'Custom JSON 2',
      value: displayContact.custom_json_2,
      show: !!displayContact.custom_json_2
    },
    {
      key: 'custom_json_3',
      label: 'Custom JSON 3',
      value: displayContact.custom_json_3,
      show: !!displayContact.custom_json_3
    },
    {
      key: 'custom_json_4',
      label: 'Custom JSON 4',
      value: displayContact.custom_json_4,
      show: !!displayContact.custom_json_4
    },
    {
      key: 'custom_json_5',
      label: 'Custom JSON 5',
      value: displayContact.custom_json_5,
      show: !!displayContact.custom_json_5
    }
  ]

  // Check if there are any JSON fields to display
  const hasJsonFields = jsonFields.some((field) => field.show)

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
      width="90%"
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
            type: 'primary',
            ghost: true,
            buttonContent: 'Update'
          }}
        />
      }
    >
      <div className="flex h-full">
        {/* Left column - Contact Details (1/4 width) */}
        <div className="w-1/3 bg-gray-50 overflow-y-auto h-full">
          {/* Contact info at the top */}
          <div className="p-6 pb-4 border-b border-gray-200 flex flex-col items-center text-center">
            <Title level={4} style={{ margin: 0, marginBottom: '4px' }}>
              {fullName}
            </Title>
            <Text type="secondary">{displayContact.email}</Text>
          </div>

          <div className="contact-details">
            {isLoadingContact && (
              <div className="mb-4 p-2 bg-blue-50 text-blue-600 rounded text-center">
                <Spin size="small" className="mr-2" />
                <span>Refreshing contact data...</span>
              </div>
            )}

            {/* Display fields in a side-by-side layout */}
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
                  field.key !== 'last_name' &&
                  // Skip JSON fields as they'll be shown separately
                  !field.key.startsWith('custom_json_')
              )
              .map((field) => (
                <div
                  key={field.key}
                  className="py-2 px-4 grid grid-cols-2 text-xs gap-1 border-b border-dashed border-gray-300"
                >
                  <Tooltip title={`Field ID: ${field.key}`}>
                    <span className="font-semibold text-slate-600">{field.label}</span>
                  </Tooltip>
                  <span>{formatValue(field.value)}</span>
                </div>
              ))}

            {/* Custom JSON fields */}
            {hasJsonFields && (
              <div>
                {jsonFields
                  .filter((field) => field.show)
                  .map((field) => (
                    <div
                      key={field.key}
                      className="py-2 px-4 grid grid-cols-2 text-xs gap-1 border-b border-dashed border-gray-300"
                    >
                      <Tooltip title={`Field ID: ${field.key}`}>
                        <span className="font-semibold text-slate-600">{field.label}</span>
                      </Tooltip>
                      {formatJson(field.value)}
                    </div>
                  ))}
              </div>
            )}
          </div>
        </div>

        {/* Right column - Message History (3/4 width) */}
        <div className="w-2/3 p-6 overflow-y-auto h-full">
          {/* E-commerce Stats (3-column grid) */}
          <div className="grid grid-cols-3 gap-4 mb-6">
            {/* Lifetime Value */}
            <Tooltip
              title={
                displayContact.lifetime_value
                  ? formatCurrency(displayContact.lifetime_value)
                  : '$0.00'
              }
            >
              <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
                <div className="text-sm text-gray-500 mb-2">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faMoneyBillWave} className="mr-2" />
                    Lifetime Value
                  </span>
                </div>
                <div className="text-2xl font-semibold">
                  {formatAverage(displayContact.lifetime_value || 0)}
                </div>
              </div>
            </Tooltip>

            {/* Orders Count */}
            <Tooltip title={`${formatNumber(displayContact.orders_count || 0)} orders`}>
              <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
                <div className="text-sm text-gray-500 mb-2">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faShoppingCart} className="mr-2" />
                    Orders Count
                  </span>
                </div>
                <div className="text-2xl font-semibold">
                  {formatAverage(displayContact.orders_count || 0)}
                </div>
              </div>
            </Tooltip>

            {/* Last Order */}
            <Tooltip
              title={
                displayContact.last_order_at
                  ? `${dayjs(displayContact.last_order_at).format('LLLL')} in ${workspaceTimezone}`
                  : 'No orders yet'
              }
            >
              <div className="bg-white rounded-lg border border-gray-200 p-4 h-24 flex flex-col justify-between">
                <div className="text-sm text-gray-500 mb-2">
                  <span className="flex items-center cursor-help">
                    <FontAwesomeIcon icon={faCalendar} className="mr-2" />
                    Last Order
                  </span>
                </div>
                <div className="text-lg font-semibold">
                  {displayContact.last_order_at
                    ? dayjs(displayContact.last_order_at).fromNow()
                    : 'Never'}
                </div>
              </div>
            </Tooltip>
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
                width: '30%',
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
                width: '20%',
                render: (status: string) => <Tag color={getStatusColor(status)}>{status}</Tag>
              },
              {
                title: 'Subscribed on',
                dataIndex: 'created_at',
                key: 'created_at',
                width: '30%',
                render: (date: string) => {
                  if (!date) return '-'

                  return (
                    <Tooltip title={`${dayjs(date).format('LLLL')} in ${workspaceTimezone}`}>
                      <span>{dayjs(date).fromNow()}</span>
                    </Tooltip>
                  )
                }
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
