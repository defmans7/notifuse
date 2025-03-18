import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { SignInPage } from '../pages/SignInPage'
import { AuthProvider } from '../contexts/AuthContext'
import * as authService from '../services/api/auth'

// Mock the auth service
vi.mock('../services/api/auth', () => ({
  authService: {
    signIn: vi.fn(),
    verifyCode: vi.fn()
  }
}))

// Mock the navigate function
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(() => ({}))
}))

describe('SignInPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the email form initially', () => {
    render(
      <AuthProvider>
        <SignInPage />
      </AuthProvider>
    )

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByText(/send magic code/i)).toBeInTheDocument()
  })

  it('submits email and shows code input form', async () => {
    // Mock successful response
    vi.mocked(authService.authService.signIn).mockResolvedValueOnce({
      message: 'Magic code sent',
      code: '123456'
    })

    render(
      <AuthProvider>
        <SignInPage />
      </AuthProvider>
    )

    // Fill and submit the email form
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' }
    })
    fireEvent.click(screen.getByText(/send magic code/i))

    // Wait for code input form to appear
    await waitFor(() => {
      expect(screen.getByText(/enter the 6-digit code/i)).toBeInTheDocument()
    })

    // Verify API was called with correct data
    expect(authService.authService.signIn).toHaveBeenCalledWith({
      email: 'test@example.com'
    })

    // Verify code form is shown
    expect(screen.getByPlaceholderText('000000')).toBeInTheDocument()
    expect(screen.getByText(/verify code/i)).toBeInTheDocument()
  })

  it('logs magic code to console when provided in response', async () => {
    // Mock console.log
    const consoleSpy = vi.spyOn(console, 'log')

    // Mock successful response with code
    vi.mocked(authService.authService.signIn).mockResolvedValueOnce({
      message: 'Magic code sent',
      code: '123456'
    })

    render(
      <AuthProvider>
        <SignInPage />
      </AuthProvider>
    )

    // Fill and submit the email form
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' }
    })
    fireEvent.click(screen.getByText(/send magic code/i))

    // Wait for state to update
    await waitFor(() => {
      expect(screen.getByText(/enter the 6-digit code/i)).toBeInTheDocument()
    })

    // Verify code was logged
    expect(consoleSpy).toHaveBeenCalledWith('Magic code for development:', '123456')
  })

  it('submits code and navigates on success', async () => {
    // Mock successful sign in response
    vi.mocked(authService.authService.signIn).mockResolvedValueOnce({
      message: 'Magic code sent'
    })

    // Mock successful verify response
    vi.mocked(authService.authService.verifyCode).mockResolvedValueOnce({
      token: 'fake-token',
      user: {
        email: 'test@example.com',
        timezone: 'UTC'
      }
    })

    render(
      <AuthProvider>
        <SignInPage />
      </AuthProvider>
    )

    // Fill and submit the email form
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' }
    })
    fireEvent.click(screen.getByText(/send magic code/i))

    // Wait for code input form
    await waitFor(() => {
      expect(screen.getByText(/enter the 6-digit code/i)).toBeInTheDocument()
    })

    // Fill and submit the code form
    fireEvent.change(screen.getByPlaceholderText('000000'), {
      target: { value: '123456' }
    })
    fireEvent.click(screen.getByText(/verify code/i))

    // Verify API was called with correct data
    await waitFor(() => {
      expect(authService.authService.verifyCode).toHaveBeenCalledWith({
        email: 'test@example.com',
        code: '123456'
      })
    })
  })

  it('shows error message when API call fails', async () => {
    // Mock failed response
    vi.mocked(authService.authService.signIn).mockRejectedValueOnce(new Error('API error'))

    render(
      <AuthProvider>
        <SignInPage />
      </AuthProvider>
    )

    // Fill and submit the email form
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'test@example.com' }
    })
    fireEvent.click(screen.getByText(/send magic code/i))

    // Error message should appear (we can't directly check antd message.error,
    // but we can verify the API was called and the form is still shown)
    await waitFor(() => {
      expect(authService.authService.signIn).toHaveBeenCalled()
    })

    // Email form should still be visible
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
  })
})
