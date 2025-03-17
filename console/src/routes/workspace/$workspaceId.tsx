import { Card, Row, Col, Statistic, Button, Typography } from 'antd'
import { useParams } from '@tanstack/react-router'
import { Link } from '@tanstack/react-router'
import { MailOutlined, SettingOutlined, TeamOutlined } from '@ant-design/icons'
import { createFileRoute } from '@tanstack/react-router'

const { Title } = Typography

function WorkspaceDashboard() {
  const { workspaceId } = useParams({ from: '/workspace/$workspaceId' })

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>Workspace Dashboard</Title>

      <Row gutter={[16, 16]}>
        {/* Quick Actions */}
        <Col span={24}>
          <Card title="Quick Actions">
            <Row gutter={[16, 16]}>
              <Col xs={24} sm={8}>
                <Link to="/workspace/$workspaceId/templates" params={{ workspaceId }}>
                  <Button type="primary" icon={<MailOutlined />} block>
                    Manage Templates
                  </Button>
                </Link>
              </Col>
              <Col xs={24} sm={8}>
                <Button icon={<TeamOutlined />} block>
                  Team Members
                </Button>
              </Col>
              <Col xs={24} sm={8}>
                <Button icon={<SettingOutlined />} block>
                  Settings
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>

        {/* Statistics */}
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Total Templates" value={0} prefix={<MailOutlined />} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Team Members" value={1} prefix={<TeamOutlined />} />
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card>
            <Statistic title="Campaigns Sent" value={0} />
          </Card>
        </Col>

        {/* Recent Activity */}
        <Col span={24}>
          <Card title="Recent Activity">
            <p style={{ textAlign: 'center', color: '#999' }}>No recent activity to display</p>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export const Route = createFileRoute('/workspace/$workspaceId')({
  component: WorkspaceDashboard
})
