import { Card, Col, Row, Typography, Button, Spin, message } from 'antd'
import { useAuth } from '../contexts/AuthContext'
import { Link, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
import { PlusOutlined } from '@ant-design/icons'
import { createFileRoute } from '@tanstack/react-router'

const { Title } = Typography
const PUBLIC_ROUTES = ['/login', '/accept-invitation', '/logout']

export const Route = createFileRoute('/')({
  component: Index
})

function Index() {
  const { workspaces, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const shouldSignIn = !isAuthenticated && !PUBLIC_ROUTES.includes(window.location.pathname)

  useEffect(() => {
    if (isAuthenticated && workspaces.length === 0) {
      navigate({ to: '/workspace/create' })
    }
    if (shouldSignIn) {
      console.log('shouldSignIn', shouldSignIn, workspaces)
      navigate({ to: '/signin' })
    }
  }, [isAuthenticated, workspaces.length, navigate, shouldSignIn])

  if (shouldSignIn) {
    return null // Will redirect in useEffect
  }

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 24
        }}
      >
        <Title level={2} style={{ margin: 0 }}>
          Your Workspaces
        </Title>
        <div>
          <Button
            onClick={() => navigate({ to: '/workspace/create' })}
            type="primary"
            icon={<PlusOutlined />}
            style={{ marginRight: 12 }}
          >
            New Workspace
          </Button>
        </div>
      </div>
      <Row gutter={[16, 16]}>
        {workspaces.map((workspace) => (
          <Col key={workspace.id} xs={24} sm={12} md={8} lg={6}>
            <Card
              hoverable
              style={{ height: '100%' }}
              cover={
                workspace.settings.logo_url && (
                  <div
                    style={{
                      height: 140,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      padding: '24px',
                      background: '#f5f5f5'
                    }}
                  >
                    <img
                      alt={workspace.settings.name}
                      src={workspace.settings.logo_url}
                      style={{ maxWidth: '100%', maxHeight: '100%', objectFit: 'contain' }}
                    />
                  </div>
                )
              }
            >
              <Card.Meta
                title={workspace.settings.name || 'Unnamed Workspace'}
                description={workspace.settings.url}
              />
              <div style={{ marginTop: 16 }}>
                <Link
                  to="/workspace/$workspaceId"
                  params={{ workspaceId: workspace.id }}
                  style={{ width: '100%', display: 'block' }}
                >
                  <Button type="primary" block>
                    Open Workspace
                  </Button>
                </Link>
              </div>
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  )
}
