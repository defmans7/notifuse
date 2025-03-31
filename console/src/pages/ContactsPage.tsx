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

const filterFields: FilterField[] = [
  { key: 'email', label: 'Email', type: 'string' as const },
  { key: 'external_id', label: 'External ID', type: 'string' as const },
  { key: 'first_name', label: 'First Name', type: 'string' as const },
  { key: 'last_name', label: 'Last Name', type: 'string' as const },
  { key: 'phone', label: 'Phone', type: 'string' as const },
  { key: 'language', label: 'Language', type: 'string' as const, options: Languages },
  { key: 'country', label: 'Country', type: 'string' as const, options: CountriesFormOptions }
]

export function ContactsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/contacts' })
  const search = useSearch({ from: workspaceContactsRoute.id })
  const navigate = useNavigate()

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
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_, record) => (
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
  ]

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
    </div>
  )
}
