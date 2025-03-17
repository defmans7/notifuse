import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { createWrapper } from '../test/setup'
import { SignIn } from '../routes/signin'
import { Logout } from '../routes/logout'
import { message } from 'antd'

// Mock the auth service
vi.mock('../services/api/auth', () => ({
  authService: {
    signIn: vi.fn(),
    verifyCode: vi.fn()
  }
}))

// Mock the router navigation and route creation
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  createFileRoute: () => () => ({ component: vi.fn() })
}))

// Mock Ant Design message
vi.mock('antd', async () => {
  const actual = await vi.importActual('antd')
  return {
    ...actual,
    message: {
      success: vi.fn(),
      error: vi.fn()
    }
  }
})

// Mock fetch for workspace calls
vi.stubGlobal(
  'fetch',
  vi.fn(() =>
    Promise.resolve({
      ok: true,
      json: () => Promise.resolve([])
    })
  )
)

describe('Authentication Flow', () => {
  const wrapper = createWrapper()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Sign In', () => {
    it('should show email input form initially', () => {
      render(<SignIn />, { wrapper })

      expect(screen.getByPlaceholderText('Email')).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /send magic code/i })).toBeInTheDocument()
    })

    it('should show code input form after submitting email', async () => {
      const { authService } = await import('../services/api/auth')
      vi.mocked(authService.signIn).mockResolvedValueOnce({ message: 'Code sent' })

      render(<SignIn />, { wrapper })

      // Fill and submit email form
      await userEvent.type(screen.getByPlaceholderText('Email'), 'test@example.com')
      await userEvent.click(screen.getByRole('button', { name: /send magic code/i }))

      // Wait for code input form to appear
      await waitFor(() => {
        expect(screen.getByPlaceholderText('000000')).toBeInTheDocument()
        expect(screen.getByRole('button', { name: /verify code/i })).toBeInTheDocument()
      })

      // Verify success message was shown
      expect(message.success).toHaveBeenCalledWith('Magic code sent to your email')
    })

    it('should handle successful sign in', async () => {
      const { authService } = await import('../services/api/auth')
      vi.mocked(authService.signIn).mockResolvedValueOnce({ message: 'Code sent' })
      vi.mocked(authService.verifyCode).mockResolvedValueOnce({
        token: 'test-token',
        user: { email: 'test@example.com', timezone: 'UTC' }
      })

      render(<SignIn />, { wrapper })

      // Fill and submit email form
      await userEvent.type(screen.getByPlaceholderText('Email'), 'test@example.com')
      await userEvent.click(screen.getByRole('button', { name: /send magic code/i }))

      // Wait for code input form and submit code
      await waitFor(() => {
        expect(screen.getByPlaceholderText('000000')).toBeInTheDocument()
      })

      await userEvent.type(screen.getByPlaceholderText('000000'), '123456')
      await userEvent.click(screen.getByRole('button', { name: /verify code/i }))

      // Verify that auth token is stored
      await waitFor(() => {
        expect(localStorage.getItem('auth_token')).toBe('test-token')
        expect(JSON.parse(localStorage.getItem('user') || '{}')).toEqual({
          email: 'test@example.com',
          timezone: 'UTC'
        })
      })

      // Verify success messages were shown
      expect(message.success).toHaveBeenCalledWith('Magic code sent to your email')
      expect(message.success).toHaveBeenCalledWith('Successfully signed in')
    })

    it('should handle sign in error', async () => {
      const { authService } = await import('../services/api/auth')
      vi.mocked(authService.signIn).mockRejectedValueOnce(new Error('Failed to send code'))

      render(<SignIn />, { wrapper })

      // Fill and submit email form
      await userEvent.type(screen.getByPlaceholderText('Email'), 'test@example.com')
      await userEvent.click(screen.getByRole('button', { name: /send magic code/i }))

      // Verify error message was shown
      await waitFor(() => {
        expect(message.error).toHaveBeenCalledWith('Failed to send code')
      })
    })
  })

  describe('Logout', () => {
    it('should clear auth data and redirect to sign in', async () => {
      // Set up initial auth state
      localStorage.setItem('auth_token', 'test-token')
      localStorage.setItem('user', JSON.stringify({ email: 'test@example.com', timezone: 'UTC' }))

      render(<Logout />, { wrapper })

      // Verify that auth data is cleared
      await waitFor(() => {
        expect(localStorage.getItem('auth_token')).toBeNull()
        expect(localStorage.getItem('user')).toBeNull()
      })
    })
  })
})
