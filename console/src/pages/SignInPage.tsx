import { Form, Input, Button, Card, App, Space } from 'antd'
import { MailOutlined } from '@ant-design/icons'
import { useAuth } from '../contexts/AuthContext'
import { useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { authService } from '../services/api/auth'
import { SignInRequest, VerifyCodeRequest } from '../services/api/types'

export function SignInPage() {
  const { signin } = useAuth()
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [showCodeInput, setShowCodeInput] = useState(false)
  const [loading, setLoading] = useState(false)
  const [resendLoading, setResendLoading] = useState(false)
  const { message } = App.useApp()

  const handleEmailSubmit = async (values: SignInRequest) => {
    try {
      setLoading(true)
      const response = await authService.signIn(values)

      // Log code if present (for development)
      if (response.code && response.code !== '') {
        console.log('Magic code for development:', response.code)
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
      const response = await authService.signIn({ email })

      // Log code if present (for development)
      if (response.code) {
        console.log('⚡ Magic code for development:', response.code)
      }

      message.success('New magic code sent to your email')
    } catch (error) {
      message.error('Failed to resend magic code')
    } finally {
      setResendLoading(false)
    }
  }

  const handleCodeSubmit = async (values: { code: string }) => {
    try {
      setLoading(true)
      const data: VerifyCodeRequest = {
        email,
        code: values.code
      }

      const response = await authService.verifyCode(data)
      const { token } = response
      // Use the existing signin function for now
      // This might need to be updated in AuthContext
      await signin(token)
      message.success('Successfully signed in')

      // Add a small delay to ensure auth state is updated before navigation
      setTimeout(() => {
        navigate({ to: '/' })
      }, 100)
    } catch (error) {
      message.error('Failed to verify code')
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
      <Card title="Sign In" style={{ width: 400 }}>
        {!showCodeInput ? (
          <Form name="email" onFinish={handleEmailSubmit} layout="vertical">
            <Form.Item
              label="Email"
              name="email"
              rules={[
                { required: true, message: 'Please input your email!' },
                { type: 'email', message: 'Please enter a valid email!' }
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="Email" type="email" />
            </Form.Item>

            <Form.Item>
              <Button type="primary" htmlType="submit" block loading={loading}>
                Send Magic Code
              </Button>
            </Form.Item>
          </Form>
        ) : (
          <>
            <p style={{ marginBottom: 24 }}>Enter the 6-digit code sent to {email}</p>
            <Form name="code" onFinish={handleCodeSubmit} layout="vertical">
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
                <Button type="primary" htmlType="submit" block loading={loading}>
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
