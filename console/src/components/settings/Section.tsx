import { Row, Col, Divider } from 'antd'

interface SectionProps {
  title: string
  description: string
  children: React.ReactNode
}

export function Section({ title, description, children }: SectionProps) {
  return (
    <div className="mb-12">
      <Row gutter={64}>
        <Col span={8}>
          <div className="text-lg font-medium">{title}</div>
          <div className="mb-6 text-sm text-gray-500">{description}</div>
        </Col>
        <Col span={16}>{children}</Col>
      </Row>
      <Divider />
    </div>
  )
}
