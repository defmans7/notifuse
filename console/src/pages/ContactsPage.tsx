import { useQuery } from '@tanstack/react-query'
import {
  Table,
  Tag,
  Button,
  Drawer,
  Form,
  Input,
  Space,
  Select,
  Typography,
  Divider,
  Alert,
  InputNumber,
  DatePicker
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useParams, useSearch, useNavigate } from '@tanstack/react-router'
import { contactsApi, type Contact, type ListContactsRequest } from '../services/api/contacts'
import React from 'react'
import { contactsRoute } from '../router'
import { Filter } from '../components/filters/Filter'
import { PlusOutlined } from '@ant-design/icons'

const { Option } = Select
const { Text } = Typography

const filterFields = [
  { key: 'email', label: 'Email', type: 'string' as const },
  { key: 'firstName', label: 'First Name', type: 'string' as const },
  { key: 'lastName', label: 'Last Name', type: 'string' as const },
  { key: 'phone', label: 'Phone', type: 'string' as const },
  { key: 'country', label: 'Country', type: 'string' as const },
  { key: 'externalId', label: 'External ID', type: 'string' as const }
]

const optionalFields = [
  { key: 'firstName', label: 'First Name' },
  { key: 'lastName', label: 'Last Name' },
  { key: 'phone', label: 'Phone' },
  { key: 'country', label: 'Country' },
  { key: 'externalId', label: 'External ID' },
  { key: 'timezone', label: 'Timezone' },
  { key: 'language', label: 'Language' },
  { key: 'addressLine1', label: 'Address Line 1' },
  { key: 'addressLine2', label: 'Address Line 2' },
  { key: 'postcode', label: 'Postcode' },
  { key: 'state', label: 'State' },
  { key: 'jobTitle', label: 'Job Title' },
  { key: 'lifetimeValue', label: 'Lifetime Value' },
  { key: 'ordersCount', label: 'Orders Count' },
  { key: 'lastOrderAt', label: 'Last Order At' },
  { key: 'customString1', label: 'Custom String 1' },
  { key: 'customString2', label: 'Custom String 2' },
  { key: 'customString3', label: 'Custom String 3' },
  { key: 'customString4', label: 'Custom String 4' },
  { key: 'customString5', label: 'Custom String 5' },
  { key: 'customNumber1', label: 'Custom Number 1' },
  { key: 'customNumber2', label: 'Custom Number 2' },
  { key: 'customNumber3', label: 'Custom Number 3' },
  { key: 'customNumber4', label: 'Custom Number 4' },
  { key: 'customNumber5', label: 'Custom Number 5' },
  { key: 'customDatetime1', label: 'Custom Date 1' },
  { key: 'customDatetime2', label: 'Custom Date 2' },
  { key: 'customDatetime3', label: 'Custom Date 3' },
  { key: 'customDatetime4', label: 'Custom Date 4' },
  { key: 'customDatetime5', label: 'Custom Date 5' }
]

export function ContactsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/contacts' })
  const search = useSearch({ from: contactsRoute.id })
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [drawerVisible, setDrawerVisible] = React.useState(false)
  const [selectedFields, setSelectedFields] = React.useState<string[]>([])
  const [selectedFieldToAdd, setSelectedFieldToAdd] = React.useState<string | null>(null)

  const activeFilters = React.useMemo(() => {
    return Object.entries(search)
      .filter(
        ([key, value]) =>
          filterFields.some((field) => field.key === key) && value !== undefined && value !== ''
      )
      .map(([key, value]) => {
        const field = filterFields.find((f) => f.key === key)
        return {
          field: key,
          value,
          label: field?.label || key
        }
      })
  }, [search])

  const { data, isLoading } = useQuery({
    queryKey: ['contacts', workspaceId, search],
    queryFn: () => {
      const request: ListContactsRequest = {
        workspaceId,
        cursor: search.cursor,
        limit: search.limit || 20,
        email: search.email,
        externalId: search.externalId,
        firstName: search.firstName,
        lastName: search.lastName,
        phone: search.phone,
        country: search.country
      }
      return contactsApi.list(request)
    }
  })

  const columns: ColumnsType<Contact> = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email'
    },
    {
      title: 'Name',
      key: 'name',
      render: (_, record) => `${record.first_name} ${record.last_name}`
    },
    {
      title: 'Country Code',
      dataIndex: 'country_code',
      key: 'country_code'
    },
    {
      title: 'Subscriptions',
      key: 'subscriptions',
      render: (_, record) => (
        <>
          {record.subscriptions.map((subscription) => (
            <Tag key={subscription.id} color="blue">
              {subscription.name}
            </Tag>
          ))}
        </>
      )
    }
  ]

  const handleAddField = () => {
    if (selectedFieldToAdd && !selectedFields.includes(selectedFieldToAdd)) {
      setSelectedFields([...selectedFields, selectedFieldToAdd])
      setSelectedFieldToAdd(null)
    }
  }

  const handleRemoveField = (field: string) => {
    setSelectedFields(selectedFields.filter((f) => f !== field))
    form.setFieldValue(field, undefined)
  }

  const handleSubmit = (values: any) => {
    console.log('Submitted values:', values)
    // Add the workspace ID to the contact data
    const contactData = {
      ...values,
      workspace_id: workspaceId
    }

    // Here you would implement the API call to create/update contact
    console.log('Contact data to send:', contactData)

    setDrawerVisible(false)
    form.resetFields()
    setSelectedFields([])
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Contacts</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setDrawerVisible(true)}>
          Insert / Update Contact
        </Button>
      </div>

      <Filter fields={filterFields} activeFilters={activeFilters} className="mb-6" />
      <Table
        columns={columns}
        dataSource={data?.contacts}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: search.cursor ? 2 : 1,
          onChange: (page) => {
            if (page > 1 && data?.next_cursor) {
              navigate({
                to: contactsRoute.id,
                search: {
                  ...search,
                  cursor: data.next_cursor
                },
                params: { workspaceId },
                replace: true
              })
            } else {
              navigate({
                to: contactsRoute.id,
                search: {
                  ...search,
                  cursor: undefined
                },
                params: { workspaceId },
                replace: true
              })
            }
          },
          total: data?.next_cursor ? (data.contacts.length || 0) + 1 : data?.contacts.length,
          pageSize: search.limit || 20
        }}
      />

      <Drawer
        title="Insert / Update Contact"
        width={500}
        open={drawerVisible}
        onClose={() => {
          setDrawerVisible(false)
          form.resetFields()
          setSelectedFields([])
        }}
        extra={
          <Space>
            <Button
              onClick={() => {
                setDrawerVisible(false)
                form.resetFields()
                setSelectedFields([])
              }}
            >
              Cancel
            </Button>
            <Button type="primary" onClick={() => form.submit()}>
              Save
            </Button>
          </Space>
        }
      >
        <Alert
          message="Note"
          description="If a contact with this email already exists, the provided fields will be overwritten. Fields not included in the form will remain unchanged."
          type="info"
          showIcon
          style={{ marginBottom: '16px' }}
        />
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="email"
            label="Email"
            rules={[
              { required: true, message: 'Email is required' },
              { type: 'email', message: 'Please enter a valid email' }
            ]}
          >
            <Input placeholder="Enter email address" />
          </Form.Item>

          {selectedFields.map((field) => {
            const fieldInfo = optionalFields.find((f) => f.key === field)

            // Render appropriate input based on field type
            let inputComponent
            if (
              field === 'lifetimeValue' ||
              field === 'ordersCount' ||
              field === 'customNumber1' ||
              field === 'customNumber2' ||
              field === 'customNumber3' ||
              field === 'customNumber4' ||
              field === 'customNumber5'
            ) {
              inputComponent = (
                <InputNumber
                  placeholder={`Enter ${fieldInfo?.label?.toLowerCase()}`}
                  style={{ width: '100%' }}
                />
              )
            } else if (
              field === 'lastOrderAt' ||
              field === 'customDatetime1' ||
              field === 'customDatetime2' ||
              field === 'customDatetime3' ||
              field === 'customDatetime4' ||
              field === 'customDatetime5'
            ) {
              inputComponent = <DatePicker showTime style={{ width: '100%' }} />
            } else if (field === 'timezone') {
              inputComponent = (
                <Select placeholder="Select timezone" style={{ width: '100%' }}>
                  <Option value="UTC">UTC</Option>
                  <Option value="America/New_York">Eastern Time (ET)</Option>
                  <Option value="America/Chicago">Central Time (CT)</Option>
                  <Option value="America/Denver">Mountain Time (MT)</Option>
                  <Option value="America/Los_Angeles">Pacific Time (PT)</Option>
                  <Option value="Europe/London">London</Option>
                  <Option value="Europe/Paris">Paris</Option>
                  <Option value="Asia/Tokyo">Tokyo</Option>
                </Select>
              )
            } else if (field === 'language') {
              inputComponent = (
                <Select placeholder="Select language" style={{ width: '100%' }}>
                  <Option value="en">English</Option>
                  <Option value="es">Spanish</Option>
                  <Option value="fr">French</Option>
                  <Option value="de">German</Option>
                  <Option value="it">Italian</Option>
                  <Option value="pt">Portuguese</Option>
                  <Option value="ja">Japanese</Option>
                  <Option value="zh">Chinese</Option>
                </Select>
              )
            } else {
              inputComponent = <Input placeholder={`Enter ${fieldInfo?.label?.toLowerCase()}`} />
            }

            return (
              <Form.Item
                key={field}
                name={field}
                label={
                  <Space>
                    <span>{fieldInfo?.label}</span>
                    <Button
                      type="text"
                      size="small"
                      danger
                      onClick={() => handleRemoveField(field)}
                      style={{ marginLeft: 'auto' }}
                    >
                      Remove
                    </Button>
                  </Space>
                }
              >
                {inputComponent}
              </Form.Item>
            )
          })}

          <Divider />

          <div>
            <Text strong>Add an optional field</Text>
            <div className="mt-2 flex gap-2">
              <Select
                placeholder="Select a field"
                style={{ width: '80%' }}
                value={selectedFieldToAdd}
                onChange={setSelectedFieldToAdd}
              >
                {optionalFields
                  .filter((field) => !selectedFields.includes(field.key))
                  .map((field) => (
                    <Option key={field.key} value={field.key}>
                      {field.label}
                    </Option>
                  ))}
              </Select>
              <Button type="primary" onClick={handleAddField} disabled={!selectedFieldToAdd}>
                Add
              </Button>
            </div>
          </div>
        </Form>
      </Drawer>
    </div>
  )
}
