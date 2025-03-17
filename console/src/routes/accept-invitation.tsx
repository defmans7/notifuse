import { Button, Card, Form, Input } from 'antd'
import { LockOutlined } from '@ant-design/icons'
import { createFileRoute } from '@tanstack/react-router'

interface AcceptInvitationForm {
  password: string
  confirmPassword: string
}

function AcceptInvitation() {
  const onFinish = (values: AcceptInvitationForm) => {
    console.log('Accept invitation form submitted:', values)
  }

  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh'
      }}
    >
      <Card title="Accept Invitation" style={{ width: 400 }}>
        <Form
          name="accept-invitation"
          onFinish={onFinish}
          validateMessages={{
            required: '${label} is required'
          }}
        >
          <Form.Item
            name="password"
            label="Password"
            rules={[
              { required: true },
              { min: 8, message: 'Password must be at least 8 characters' }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>

          <Form.Item
            name="confirmPassword"
            label="Confirm Password"
            dependencies={['password']}
            rules={[
              { required: true },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('password') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject('The passwords do not match')
                }
              })
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" style={{ width: '100%' }}>
              Set Password & Accept
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  )
}

export const Route = createFileRoute('/accept-invitation')({
  component: AcceptInvitation
})
