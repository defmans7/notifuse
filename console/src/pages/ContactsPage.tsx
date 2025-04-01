import { useQuery } from '@tanstack/react-query'
import { Table, Tag, Button } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useParams, useSearch, useNavigate } from '@tanstack/react-router'
import { contactsApi, type Contact, type ListContactsRequest } from '../services/api/contacts'
import React from 'react'
import { workspaceContactsRoute } from '../router'
import { Filter } from '../components/filters/Filter'
import { ContactUpsertDrawer } from '../components/contacts/ContactUpsertDrawer'
import { CountriesFormOptions } from '../components/utils/countries_timezones'
import { Languages } from '../components/utils/languages'
import { FilterField } from '../components/filters/types'
import { ContactColumnsSelector, JsonViewer } from '../components/contacts/ContactColumnsSelector'

const filterFields: FilterField[] = [
  { key: 'email', label: 'Email', type: 'string' as const },
  { key: 'external_id', label: 'External ID', type: 'string' as const },
  { key: 'first_name', label: 'First Name', type: 'string' as const },
  { key: 'last_name', label: 'Last Name', type: 'string' as const },
  { key: 'phone', label: 'Phone', type: 'string' as const },
  { key: 'language', label: 'Language', type: 'string' as const, options: Languages },
  { key: 'country', label: 'Country', type: 'string' as const, options: CountriesFormOptions }
]

const STORAGE_KEY = 'contact_columns_visibility'

const DEFAULT_VISIBLE_COLUMNS = {
  name: true,
  language: true,
  timezone: true,
  country: true,
  lists: true,
  phone: false,
  address: false,
  job_title: false,
  lifetime_value: false,
  orders_count: false,
  last_order_at: false,
  custom_string_1: false,
  custom_string_2: false,
  custom_string_3: false,
  custom_string_4: false,
  custom_string_5: false,
  custom_number_1: false,
  custom_number_2: false,
  custom_number_3: false,
  custom_number_4: false,
  custom_number_5: false,
  custom_datetime_1: false,
  custom_datetime_2: false,
  custom_datetime_3: false,
  custom_datetime_4: false,
  custom_datetime_5: false,
  custom_json_1: false,
  custom_json_2: false,
  custom_json_3: false,
  custom_json_4: false,
  custom_json_5: false
}

export function ContactsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/contacts' })
  const search = useSearch({ from: workspaceContactsRoute.id })
  const navigate = useNavigate()

  const [visibleColumns, setVisibleColumns] =
    React.useState<Record<string, boolean>>(DEFAULT_VISIBLE_COLUMNS)

  // Load saved state from localStorage on mount
  React.useEffect(() => {
    const savedState = localStorage.getItem(STORAGE_KEY)
    if (savedState) {
      const parsedState = JSON.parse(savedState)
      // Merge with defaults to ensure all fields exist
      setVisibleColumns({
        ...DEFAULT_VISIBLE_COLUMNS,
        ...parsedState
      })
    }
  }, [])

  const handleColumnVisibilityChange = (key: string, visible: boolean) => {
    setVisibleColumns((prev) => {
      const newState = { ...prev, [key]: visible }
      // Save to localStorage
      localStorage.setItem(STORAGE_KEY, JSON.stringify(newState))
      return newState
    })
  }

  const allColumns: { key: string; title: string }[] = [
    { key: 'name', title: 'Name' },
    { key: 'phone', title: 'Phone' },
    { key: 'country', title: 'Country' },
    { key: 'language', title: 'Language' },
    { key: 'timezone', title: 'Timezone' },
    { key: 'address', title: 'Address' },
    { key: 'job_title', title: 'Job Title' },
    { key: 'lifetime_value', title: 'Lifetime Value' },
    { key: 'orders_count', title: 'Orders Count' },
    { key: 'last_order_at', title: 'Last Order' },
    { key: 'lists', title: 'Lists' },
    { key: 'custom_string_1', title: 'Custom String 1' },
    { key: 'custom_string_2', title: 'Custom String 2' },
    { key: 'custom_string_3', title: 'Custom String 3' },
    { key: 'custom_string_4', title: 'Custom String 4' },
    { key: 'custom_string_5', title: 'Custom String 5' },
    { key: 'custom_number_1', title: 'Custom Number 1' },
    { key: 'custom_number_2', title: 'Custom Number 2' },
    { key: 'custom_number_3', title: 'Custom Number 3' },
    { key: 'custom_number_4', title: 'Custom Number 4' },
    { key: 'custom_number_5', title: 'Custom Number 5' },
    { key: 'custom_datetime_1', title: 'Custom Date 1' },
    { key: 'custom_datetime_2', title: 'Custom Date 2' },
    { key: 'custom_datetime_3', title: 'Custom Date 3' },
    { key: 'custom_datetime_4', title: 'Custom Date 4' },
    { key: 'custom_datetime_5', title: 'Custom Date 5' },
    { key: 'custom_json_1', title: 'Custom JSON 1' },
    { key: 'custom_json_2', title: 'Custom JSON 2' },
    { key: 'custom_json_3', title: 'Custom JSON 3' },
    { key: 'custom_json_4', title: 'Custom JSON 4' },
    { key: 'custom_json_5', title: 'Custom JSON 5' }
  ]

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
      key: 'email',
      fixed: 'left' as const,
      onHeaderCell: () => ({
        className: '!bg-white'
      })
    },
    {
      title: 'Name',
      key: 'name',
      render: (_: unknown, record: Contact) =>
        `${record.first_name || ''} ${record.last_name || ''}`,
      hidden: !visibleColumns.name
    },
    {
      title: 'Phone',
      dataIndex: 'phone',
      key: 'phone',
      hidden: !visibleColumns.phone
    },
    {
      title: 'Country',
      dataIndex: 'country',
      key: 'country',
      hidden: !visibleColumns.country
    },
    {
      title: 'Language',
      dataIndex: 'language',
      key: 'language',
      hidden: !visibleColumns.language
    },
    {
      title: 'Timezone',
      dataIndex: 'timezone',
      key: 'timezone',
      hidden: !visibleColumns.timezone
    },
    {
      title: 'Address',
      key: 'address',
      render: (_: unknown, record: Contact) => {
        const parts = [
          record.address_line_1,
          record.address_line_2,
          record.state,
          record.postcode
        ].filter(Boolean)
        return parts.join(', ')
      },
      hidden: !visibleColumns.address
    },
    {
      title: 'Job Title',
      dataIndex: 'job_title',
      key: 'job_title',
      hidden: !visibleColumns.job_title
    },
    {
      title: 'Lifetime Value',
      dataIndex: 'lifetime_value',
      key: 'lifetime_value',
      render: (_: unknown, record: Contact) =>
        record.lifetime_value ? `$${record.lifetime_value.toFixed(2)}` : '-',
      hidden: !visibleColumns.lifetime_value
    },
    {
      title: 'Orders Count',
      dataIndex: 'orders_count',
      key: 'orders_count',
      hidden: !visibleColumns.orders_count
    },
    {
      title: 'Last Order',
      dataIndex: 'last_order_at',
      key: 'last_order_at',
      render: (_: unknown, record: Contact) =>
        record.last_order_at ? new Date(record.last_order_at).toLocaleDateString() : '-',
      hidden: !visibleColumns.last_order_at
    },
    {
      title: 'Custom String 1',
      dataIndex: 'custom_string_1',
      key: 'custom_string_1',
      hidden: !visibleColumns.custom_string_1
    },
    {
      title: 'Custom String 2',
      dataIndex: 'custom_string_2',
      key: 'custom_string_2',
      hidden: !visibleColumns.custom_string_2
    },
    {
      title: 'Custom String 3',
      dataIndex: 'custom_string_3',
      key: 'custom_string_3',
      hidden: !visibleColumns.custom_string_3
    },
    {
      title: 'Custom String 4',
      dataIndex: 'custom_string_4',
      key: 'custom_string_4',
      hidden: !visibleColumns.custom_string_4
    },
    {
      title: 'Custom String 5',
      dataIndex: 'custom_string_5',
      key: 'custom_string_5',
      hidden: !visibleColumns.custom_string_5
    },
    {
      title: 'Custom Number 1',
      dataIndex: 'custom_number_1',
      key: 'custom_number_1',
      hidden: !visibleColumns.custom_number_1
    },
    {
      title: 'Custom Number 2',
      dataIndex: 'custom_number_2',
      key: 'custom_number_2',
      hidden: !visibleColumns.custom_number_2
    },
    {
      title: 'Custom Number 3',
      dataIndex: 'custom_number_3',
      key: 'custom_number_3',
      hidden: !visibleColumns.custom_number_3
    },
    {
      title: 'Custom Number 4',
      dataIndex: 'custom_number_4',
      key: 'custom_number_4',
      hidden: !visibleColumns.custom_number_4
    },
    {
      title: 'Custom Number 5',
      dataIndex: 'custom_number_5',
      key: 'custom_number_5',
      hidden: !visibleColumns.custom_number_5
    },
    {
      title: 'Custom Date 1',
      dataIndex: 'custom_datetime_1',
      key: 'custom_datetime_1',
      render: (_: unknown, record: Contact) =>
        record.custom_datetime_1 ? new Date(record.custom_datetime_1).toLocaleDateString() : '-',
      hidden: !visibleColumns.custom_datetime_1
    },
    {
      title: 'Custom Date 2',
      dataIndex: 'custom_datetime_2',
      key: 'custom_datetime_2',
      render: (_: unknown, record: Contact) =>
        record.custom_datetime_2 ? new Date(record.custom_datetime_2).toLocaleDateString() : '-',
      hidden: !visibleColumns.custom_datetime_2
    },
    {
      title: 'Custom Date 3',
      dataIndex: 'custom_datetime_3',
      key: 'custom_datetime_3',
      render: (_: unknown, record: Contact) =>
        record.custom_datetime_3 ? new Date(record.custom_datetime_3).toLocaleDateString() : '-',
      hidden: !visibleColumns.custom_datetime_3
    },
    {
      title: 'Custom Date 4',
      dataIndex: 'custom_datetime_4',
      key: 'custom_datetime_4',
      render: (_: unknown, record: Contact) =>
        record.custom_datetime_4 ? new Date(record.custom_datetime_4).toLocaleDateString() : '-',
      hidden: !visibleColumns.custom_datetime_4
    },
    {
      title: 'Custom Date 5',
      dataIndex: 'custom_datetime_5',
      key: 'custom_datetime_5',
      render: (_: unknown, record: Contact) =>
        record.custom_datetime_5 ? new Date(record.custom_datetime_5).toLocaleDateString() : '-',
      hidden: !visibleColumns.custom_datetime_5
    },
    {
      title: 'Custom JSON 1',
      dataIndex: 'custom_json_1',
      key: 'custom_json_1',
      render: (_: unknown, record: Contact) => (
        <JsonViewer json={record.custom_json_1} title="Custom JSON 1" />
      ),
      hidden: !visibleColumns.custom_json_1
    },
    {
      title: 'Custom JSON 2',
      dataIndex: 'custom_json_2',
      key: 'custom_json_2',
      render: (_: unknown, record: Contact) => (
        <JsonViewer json={record.custom_json_2} title="Custom JSON 2" />
      ),
      hidden: !visibleColumns.custom_json_2
    },
    {
      title: 'Custom JSON 3',
      dataIndex: 'custom_json_3',
      key: 'custom_json_3',
      render: (_: unknown, record: Contact) => (
        <JsonViewer json={record.custom_json_3} title="Custom JSON 3" />
      ),
      hidden: !visibleColumns.custom_json_3
    },
    {
      title: 'Custom JSON 4',
      dataIndex: 'custom_json_4',
      key: 'custom_json_4',
      render: (_: unknown, record: Contact) => (
        <JsonViewer json={record.custom_json_4} title="Custom JSON 4" />
      ),
      hidden: !visibleColumns.custom_json_4
    },
    {
      title: 'Custom JSON 5',
      dataIndex: 'custom_json_5',
      key: 'custom_json_5',
      render: (_: unknown, record: Contact) => (
        <JsonViewer json={record.custom_json_5} title="Custom JSON 5" />
      ),
      hidden: !visibleColumns.custom_json_5
    },
    {
      title: 'Lists',
      key: 'lists',
      render: (_: unknown, record: Contact) => (
        <>
          {record.contact_lists.map((list: { list_id: string }) => (
            <Tag key={list.list_id} color="blue">
              {list.list_id}
            </Tag>
          ))}
        </>
      ),
      hidden: !visibleColumns.lists
    },
    {
      title: 'Actions',
      key: 'actions',
      fixed: 'right' as const,
      render: (_: unknown, record: Contact) => (
        <ContactUpsertDrawer
          workspaceId={workspaceId}
          contact={record}
          buttonProps={{
            type: 'link',
            buttonContent: 'Edit',
            size: 'small'
          }}
        />
      )
    }
  ].filter((col) => !col.hidden)

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
        <ContactUpsertDrawer workspaceId={workspaceId} />
      </div>

      <div className="flex justify-between items-center mb-6">
        <Filter fields={filterFields} activeFilters={activeFilters} />
        <ContactColumnsSelector
          columns={allColumns.map((col) => ({
            ...col,
            visible: visibleColumns[col.key]
          }))}
          onColumnVisibilityChange={handleColumnVisibilityChange}
        />
      </div>

      <Table
        columns={columns}
        dataSource={data?.contacts}
        rowKey={(record) => record.email}
        loading={isLoading}
        pagination={false}
        scroll={{ x: 'max-content' }}
        style={{ minWidth: 800 }}
      />

      {data?.cursor && (
        <div className="flex justify-center mt-4">
          <Button onClick={handleLoadMore} loading={isLoading}>
            Load More
          </Button>
        </div>
      )}
    </div>
  )
}
