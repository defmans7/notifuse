import { render, screen, waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { Root } from '../../routes/__root'
import { useAuth } from '../../contexts/AuthContext'
import { useNavigate, useMatchRoute } from '@tanstack/react-router'

// Mock the hooks
vi.mock('../../contexts/AuthContext')
vi.mock('@tanstack/react-router')

describe('Root Component', () => {
  const mockNavigate = vi.fn()
  const mockMatchRoute = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    ;(useNavigate as any).mockReturnValue(mockNavigate)
    ;(useMatchRoute as any).mockReturnValue(mockMatchRoute)
  })

  it('should redirect to home when user has no workspaces and is in workspace route', async () => {
    // Mock auth context with user but no workspaces
    ;(useAuth as any).mockReturnValue({
      user: { email: 'test@example.com' },
      workspaces: []
    })

    // Mock route matching to simulate being in a workspace route
    mockMatchRoute.mockImplementation(({ to }) => {
      return (
        to === '/workspace/$workspaceId' ||
        to === '/workspace/$workspaceId/templates' ||
        to === '/workspace/$workspaceId/settings' ||
        to === '/workspace/$workspaceId/contacts' ||
        to === '/workspace/$workspaceId/campaigns'
      )
    })

    // Mock window.location
    const originalLocation = window.location
    delete (window as any).location
    window.location = {
      ...originalLocation,
      pathname: '/workspace/123'
    }

    render(<Root />)

    // Wait for the navigation to be called
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith({ to: '/' })
    })
  })

  it('should not redirect when user has workspaces', async () => {
    // Mock auth context with user and workspaces
    ;(useAuth as any).mockReturnValue({
      user: { email: 'test@example.com' },
      workspaces: [{ id: '123', name: 'Test Workspace' }]
    })

    // Mock route matching to simulate being in a workspace route
    mockMatchRoute.mockImplementation(({ to }) => {
      return (
        to === '/workspace/$workspaceId' ||
        to === '/workspace/$workspaceId/templates' ||
        to === '/workspace/$workspaceId/settings' ||
        to === '/workspace/$workspaceId/contacts' ||
        to === '/workspace/$workspaceId/campaigns'
      )
    })

    // Mock window.location
    const originalLocation = window.location
    delete (window as any).location
    window.location = {
      ...originalLocation,
      pathname: '/workspace/123'
    }

    render(<Root />)

    // Wait to ensure no navigation occurs
    await waitFor(() => {
      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })

  it('should not redirect when not in workspace route', async () => {
    // Mock auth context with user but no workspaces
    ;(useAuth as any).mockReturnValue({
      user: { email: 'test@example.com' },
      workspaces: []
    })

    // Mock route matching to simulate not being in a workspace route
    mockMatchRoute.mockImplementation(() => false)

    // Mock window.location
    const originalLocation = window.location
    delete (window as any).location
    window.location = {
      ...originalLocation,
      pathname: '/'
    }

    render(<Root />)

    // Wait to ensure no navigation occurs
    await waitFor(() => {
      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })

  it('should not redirect when in create workspace route', async () => {
    // Mock auth context with user but no workspaces
    ;(useAuth as any).mockReturnValue({
      user: { email: 'test@example.com' },
      workspaces: []
    })

    // Mock route matching to simulate being in create workspace route
    mockMatchRoute.mockImplementation(({ to }) => {
      return to === '/workspace/create'
    })

    // Mock window.location
    const originalLocation = window.location
    delete (window as any).location
    window.location = {
      ...originalLocation,
      pathname: '/workspace/create'
    }

    render(<Root />)

    // Wait to ensure no navigation occurs
    await waitFor(() => {
      expect(mockNavigate).not.toHaveBeenCalled()
    })
  })
})
