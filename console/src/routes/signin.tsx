import { Button, Card, Form, Input, message, Space } from 'antd'
import { MailOutlined } from '@ant-design/icons'
import { useNavigate } from '@tanstack/react-router'
import { useAuth } from '../contexts/AuthContext'
import { useState } from 'react'
import config from '../config'
import { createFileRoute } from '@tanstack/react-router'

interface EmailForm {
  email: string
}

interface MagicCodeForm {
  code: string
}

interface VerifyResponse {
  token: string
  user: {
    email: string
    timezone: string
  }
}

function SignIn() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [showCodeInput, setShowCodeInput] = useState(false)
  const [loading, setLoading] = useState(false)
  const [resendLoading, setResendLoading] = useState(false)

  const handleEmailSubmit = async (values: EmailForm) => {
    try {
      setLoading(true)
      const response = await fetch(`${config.API_ENDPOINT}/auth/magic-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email: values.email })
      })

      if (!response.ok) {
        throw new Error('Failed to send magic code')
      }

      setEmail(values.email)
      setShowCodeInput(true)
      message.success('Magic code sent to your email')
    } catch (error) {
      message.error('Failed to send magic code')
    } finally {
      setLoading(false)
    }
  }

  const handleResendCode = async () => {
    try {
      setResendLoading(true)
      const response = await fetch(`${config.API_ENDPOINT}/auth/magic-code`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email })
      })

      if (!response.ok) {
        throw new Error('Failed to resend magic code')
      }

      message.success('New magic code sent to your email')
    } catch (error) {
      message.error('Failed to resend magic code')
    } finally {
      setResendLoading(false)
    }
  }

  const handleCodeSubmit = async (values: MagicCodeForm) => {
    try {
      setLoading(true)
      const response = await fetch(`${config.API_ENDPOINT}/auth/verify`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          email,
          code: values.code
        })
      })

      if (!response.ok) {
        throw new Error('Invalid magic code')
      }

      const { token, user } = (await response.json()) as VerifyResponse

      // First login to set the token and user
      await login(token, user)
      message.success('Successfully signed in')
      navigate({ to: '/' })
    } catch (error) {
      if (error instanceof Error) {
        message.error(error.message)
      } else {
        message.error('Failed to sign in')
      }
    } finally {
      setLoading(false)
    }
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
      <Card title="Sign in to Notifuse" style={{ width: 400 }}>
        {!showCodeInput ? (
          <Form name="email" onFinish={handleEmailSubmit}>
            <Form.Item
              name="email"
              rules={[
                { required: true, message: 'Please input your email!' },
                { type: 'email', message: 'Please enter a valid email!' }
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="Email" type="email" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" style={{ width: '100%' }} loading={loading}>
                Send Magic Code
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <>
            <p style={{ marginBottom: 24 }}>Enter the 6-digit code sent to {email}</p>
            <Form name="code" onFinish={handleCodeSubmit}>
              <Form.Item
                name="code"
                rules={[
                  { required: true, message: 'Please input the magic code!' },
                  {
                    pattern: /^\d{6}$/,
                    message: 'Please enter a valid 6-digit code!'
                  }
                ]}
              >
                <Input
                  placeholder="000000"
                  maxLength={6}
                  style={{ textAlign: 'center', letterSpacing: '0.5em' }}
                />
              </Form.Item>

              <Form.Item>
                <Button
                  type="primary"
                  htmlType="submit"
                  style={{ width: '100%' }}
                  loading={loading}
                >
                  Verify Code
                </Button>
              </Form.Item>

              <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                <Button type="link" onClick={() => setShowCodeInput(false)} style={{ padding: 0 }}>
                  Use a different email
                </Button>
                <Button
                  type="link"
                  onClick={handleResendCode}
                  loading={resendLoading}
                  style={{ padding: 0 }}
                >
                  Resend code
                </Button>
              </Space>
            </Form>
          </>
        )}
      </Card>
    </div>
  )
}

export const Route = createFileRoute('/signin')({
  component: SignIn
})
