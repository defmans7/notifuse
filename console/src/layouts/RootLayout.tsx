import { Outlet, useNavigate, useMatch } from '@tanstack/react-router'
import { Spin } from 'antd'
import { useAuth } from '../contexts/AuthContext'
import { useEffect } from 'react'

export function RootLayout() {
  const { isAuthenticated, loading, workspaces } = useAuth()
  const navigate = useNavigate()

  const isSigninRoute = useMatch({ from: '/signin', shouldThrow: false })
  const isAcceptInvitationRoute = useMatch({ from: '/accept-invitation', shouldThrow: false })
  const isLogoutRoute = useMatch({ from: '/logout', shouldThrow: false })
  const isWorkspaceCreateRoute = useMatch({ from: '/workspace/create', shouldThrow: false })

  const isPublicRoute = isSigninRoute || isAcceptInvitationRoute || isLogoutRoute
  // If not authenticated and not on public routes, redirect to signin
  const shouldRedirectToSignin = !isSigninRoute && !isAuthenticated && !isPublicRoute

  // If authenticated and has no workspaces, redirect to workspace creation
  const shouldRedirectToCreateWorkspace =
    isAuthenticated && workspaces.length === 0 && !isWorkspaceCreateRoute

  // handle redirection...
  useEffect(() => {
    if (loading) return

    if (shouldRedirectToSignin) {
      navigate({ to: '/signin' })
      return
    }

    if (shouldRedirectToCreateWorkspace) {
      navigate({ to: '/workspace/create' })
      return
    }
  }, [loading, shouldRedirectToSignin, shouldRedirectToCreateWorkspace])

  if (loading || shouldRedirectToSignin || shouldRedirectToCreateWorkspace) {
    return (
      <div
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}
      >
        <Spin size="large" />
      </div>
    )
  }

  console.log('Rendering RootLayout')
  return <Outlet />
}
