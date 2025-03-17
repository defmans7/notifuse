import { Button, Table, Space } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { createFileRoute } from '@tanstack/react-router'

interface Template {
  id: string
  name: string
  subject: string
  created_at: string
  updated_at: string
}

function Templates() {
  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name'
    },
    {
      title: 'Subject',
      dataIndex: 'subject',
      key: 'subject'
    },
    {
      title: 'Created At',
      dataIndex: 'created_at',
      key: 'created_at'
    },
    {
      title: 'Updated At',
      dataIndex: 'updated_at',
      key: 'updated_at'
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (_: any, record: Template) => (
        <Space size="middle">
          <Button type="link">Edit</Button>
          <Button type="link" danger>
            Delete
          </Button>
        </Space>
      )
    }
  ]

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          marginBottom: 16
        }}
      >
        <h1>Email Templates</h1>
        <Button type="primary" icon={<PlusOutlined />}>
          New Template
        </Button>
      </div>
      <Table columns={columns} dataSource={[]} rowKey="id" />
    </div>
  )
}

export const Route = createFileRoute('/workspace/$workspaceId/templates')({
  component: Templates
})
