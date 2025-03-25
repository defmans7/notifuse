import { useQuery } from '@tanstack/react-query'
import { Table, Tag, Input, Form, Button, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { useParams, useSearch, useNavigate } from '@tanstack/react-router'
import { contactsApi, type Contact, type ListContactsRequest } from '../services/api/contacts'
import React from 'react'
import { contactsRoute, type ContactsSearch } from '../router'

export function ContactsPage() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId/contacts' })
  const search = useSearch({ from: contactsRoute.id })
  const navigate = useNavigate()
  const [form] = Form.useForm()

  // Initialize form with current search params
  React.useEffect(() => {
    form.setFieldsValue({
      email: search.email,
      externalId: search.externalId,
      firstName: search.firstName,
      lastName: search.lastName,
      phone: search.phone,
      country: search.country
    })
  }, [form, search])

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

  const handleSearch = (values: ContactsSearch) => {
    // Remove empty values and navigate
    const cleanValues = Object.fromEntries(
      Object.entries(values).filter(([_, value]) => value !== undefined && value !== '')
    ) as ContactsSearch
    navigate({
      to: contactsRoute.id,
      search: cleanValues,
      params: { workspaceId },
      replace: true
    })
  }

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

  return (
    <div className="p-6">
      <h2 className="mb-6">Contacts</h2>

      <Form form={form} layout="inline" onFinish={handleSearch} className="mb-6">
        <Form.Item name="email">
          <Input placeholder="Email" allowClear />
        </Form.Item>
        <Form.Item name="firstName">
          <Input placeholder="First Name" allowClear />
        </Form.Item>
        <Form.Item name="lastName">
          <Input placeholder="Last Name" allowClear />
        </Form.Item>
        <Form.Item name="phone">
          <Input placeholder="Phone" allowClear />
        </Form.Item>
        <Form.Item name="country">
          <Input placeholder="Country" allowClear />
        </Form.Item>
        <Form.Item name="externalId">
          <Input placeholder="External ID" allowClear />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              Search
            </Button>
            <Button
              onClick={() => {
                form.resetFields()
                navigate({
                  to: contactsRoute.id,
                  search: {},
                  params: { workspaceId },
                  replace: true
                })
              }}
            >
              Reset
            </Button>
          </Space>
        </Form.Item>
      </Form>

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
    </div>
  )
}
