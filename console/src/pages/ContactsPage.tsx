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
import { workspaceContactsRoute } from '../router'
import { Filter } from '../components/filters/Filter'
import { PlusOutlined } from '@ant-design/icons'
import { CountriesFormOptions, TimezonesFormOptions } from '../components/utils/countries_timezones'
import { Languages } from '../components/utils/languages'
import { FilterField } from '../components/filters/types'

const { Option } = Select
const { Text } = Typography

const filterFields: FilterField[] = [
  { key: 'email', label: 'Email', type: 'string' as const },
  { key: 'external_id', label: 'External ID', type: 'string' as const },
  { key: 'first_name', label: 'First Name', type: 'string' as const },
  { key: 'last_name', label: 'Last Name', type: 'string' as const },
  { key: 'phone', label: 'Phone', type: 'string' as const },
  { key: 'language', label: 'Language', type: 'string' as const, options: Languages },
  { key: 'country', label: 'Country', type: 'string' as const, options: CountriesFormOptions }
]

const optionalFields = [
  { key: 'first_name', label: 'First Name' },
  { key: 'last_name', label: 'Last Name' },
  { key: 'phone', label: 'Phone' },
  { key: 'country', label: 'Country' },
  { key: 'external_id', label: 'External ID' },
  { key: 'timezone', label: 'Timezone' },
  { key: 'language', label: 'Language' },
  { key: 'address_line_1', label: 'Address Line 1' },
  { key: 'address_line_2', label: 'Address Line 2' },
  { key: 'postcode', label: 'Postcode' },
  { key: 'state', label: 'State' },
  { key: 'job_title', label: 'Job Title' },
  { key: 'lifetime_value', label: 'Lifetime Value' },
  { key: 'orders_count', label: 'Orders Count' },
  { key: 'last_order_at', label: 'Last Order At' },
  { key: 'custom_string_1', label: 'Custom String 1' },
  { key: 'custom_string_2', label: 'Custom String 2' },
  { key: 'custom_string_3', label: 'Custom String 3' },
  { key: 'custom_string_4', label: 'Custom String 4' },
  { key: 'custom_string_5', label: 'Custom String 5' },
  { key: 'custom_number_1', label: 'Custom Number 1' },
  { key: 'custom_number_2', label: 'Custom Number 2' },
  { key: 'custom_number_3', label: 'Custom Number 3' },
  { key: 'custom_number_4', label: 'Custom Number 4' },
  { key: 'custom_number_5', label: 'Custom Number 5' },
  { key: 'custom_datetime_1', label: 'Custom Date 1' },
  { key: 'custom_datetime_2', label: 'Custom Date 2' },
  { key: 'custom_datetime_3', label: 'Custom Date 3' },
  { key: 'custom_datetime_4', label: 'Custom Date 4' },
  { key: 'custom_datetime_5', label: 'Custom Date 5' }
]

export function ContactsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/contacts' })
  const search = useSearch({ from: workspaceContactsRoute.id })
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
        workspace_id: workspaceId,
        cursor: search.cursor,
        limit: search.limit || 20,
        email: search.email,
        external_id: search.external_id,
        first_name: search.first_name,
        last_name: search.last_name,
        phone: search.phone,
        country: search.country,
        language: search.language,
        with_contact_lists: true
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
      render: (_, record) => `${record.first_name || ''} ${record.last_name || ''}`
    },
    {
      title: 'Country Code',
      dataIndex: 'country_code',
      key: 'country_code'
    },
    {
      title: 'Lists',
      key: 'lists',
      render: (_, record) => (
        <>
          {record.contact_lists.map((list) => (
            <Tag key={list.list_id} color="blue">
              {list.list_id}
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

  const handleLoadMore = () => {
    if (data?.cursor) {
      navigate({
        to: workspaceContactsRoute.id,
        search: {
          ...search,
          cursor: data.cursor
        },
        params: { workspaceId },
        replace: true
      })
    }
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold">Contacts</h2>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setDrawerVisible(true)}>
          Insert / Update
        </Button>
      </div>

      <Filter fields={filterFields} activeFilters={activeFilters} className="mb-6" />
      <Table
        columns={columns}
        dataSource={data?.contacts}
        rowKey={(record) => record.email}
        loading={isLoading}
        pagination={false}
      />

      {data?.cursor && (
        <div className="flex justify-center mt-4">
          <Button onClick={handleLoadMore} loading={isLoading}>
            Load More
          </Button>
        </div>
      )}

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
              field === 'lifetime_value' ||
              field === 'orders_count' ||
              field === 'custom_number_1' ||
              field === 'custom_number_2' ||
              field === 'custom_number_3' ||
              field === 'custom_number_4' ||
              field === 'custom_number_5'
            ) {
              inputComponent = (
                <InputNumber
                  key={field}
                  placeholder={`Enter ${fieldInfo?.label?.toLowerCase()}`}
                  style={{ width: '100%' }}
                />
              )
            } else if (
              field === 'last_order_at' ||
              field === 'custom_datetime_1' ||
              field === 'custom_datetime_2' ||
              field === 'custom_datetime_3' ||
              field === 'custom_datetime_4' ||
              field === 'custom_datetime_5'
            ) {
              inputComponent = <DatePicker showTime style={{ width: '100%' }} />
            } else if (field === 'timezone') {
              inputComponent = (
                <Select
                  key={field}
                  placeholder="Select timezone"
                  style={{ width: '100%' }}
                  options={TimezonesFormOptions}
                />
              )
            } else if (field === 'country') {
              inputComponent = (
                <Select
                  key={field}
                  placeholder="Select country"
                  style={{ width: '100%' }}
                  options={CountriesFormOptions}
                />
              )
            } else if (field === 'language') {
              inputComponent = (
                <Select
                  placeholder="Select language"
                  style={{ width: '100%' }}
                  options={Languages}
                />
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
