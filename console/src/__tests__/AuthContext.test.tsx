import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, act, waitFor } from '@testing-library/react'
import { AuthProvider, useAuth } from '../contexts/AuthContext'
import { ReactNode } from 'react'

// Create a test component that uses the auth context
const TestComponent = () => {
  const { user, isAuthenticated, signin, signout, loading } = useAuth()

  return (
    <div>
      <div data-testid="loading">{loading ? 'Loading' : 'Not Loading'}</div>
      <div data-testid="authenticated">
        {isAuthenticated ? 'Authenticated' : 'Not Authenticated'}
      </div>
      <div data-testid="user">{user ? JSON.stringify(user) : 'No User'}</div>
      <button data-testid="signin" onClick={() => signin('test@example.com', 'password')}>
        Sign In
      </button>
      <button data-testid="signout" onClick={() => signout()}>
        Sign Out
      </button>
    </div>
  )
}

const wrapper = ({ children }: { children: ReactNode }) => <AuthProvider>{children}</AuthProvider>

describe('AuthContext', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('provides initial auth state', async () => {
    render(<TestComponent />, { wrapper })

    // Initial state should be loading
    expect(screen.getByTestId('loading')).toHaveTextContent('Loading')

    // Wait for check auth to complete
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('Not Loading')
    })

    // Initial state should be not authenticated
    expect(screen.getByTestId('authenticated')).toHaveTextContent('Not Authenticated')
    expect(screen.getByTestId('user')).toHaveTextContent('No User')
  })

  it('handles signin action', async () => {
    render(<TestComponent />, { wrapper })

    // Wait for check auth to complete
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('Not Loading')
    })

    // Initial state
    expect(screen.getByTestId('authenticated')).toHaveTextContent('Not Authenticated')

    // Trigger signin
    await act(async () => {
      screen.getByTestId('signin').click()
    })

    // User should be authenticated
    expect(screen.getByTestId('authenticated')).toHaveTextContent('Authenticated')
    expect(screen.getByTestId('user')).toHaveTextContent('test@example.com')
  })

  it('handles signout action', async () => {
    render(<TestComponent />, { wrapper })

    // Wait for check auth to complete
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('Not Loading')
    })

    // Trigger signin first
    await act(async () => {
      screen.getByTestId('signin').click()
    })

    // Verify user is authenticated
    expect(screen.getByTestId('authenticated')).toHaveTextContent('Authenticated')

    // Trigger signout
    await act(async () => {
      screen.getByTestId('signout').click()
    })

    // User should be signed out
    expect(screen.getByTestId('authenticated')).toHaveTextContent('Not Authenticated')
    expect(screen.getByTestId('user')).toHaveTextContent('No User')
  })

  it('throws error when useAuth is used outside AuthProvider', () => {
    // Suppress console.error for this test
    const originalConsoleError = console.error
    console.error = vi.fn()

    // Using useAuth outside provider should throw
    expect(() => {
      render(<TestComponent />)
    }).toThrow('useAuth must be used within an AuthProvider')

    // Restore console.error
    console.error = originalConsoleError
  })
})
