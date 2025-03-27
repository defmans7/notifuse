import { createRootRoute, createRoute, Route, Router } from '@tanstack/react-router'
import { RootLayout } from './layouts/RootLayout'
import { WorkspaceLayout } from './layouts/WorkspaceLayout'
import { SignInPage } from './pages/SignInPage'
import { LogoutPage } from './pages/LogoutPage'
import { CreateWorkspacePage } from './pages/CreateWorkspacePage'
import { DashboardPage } from './pages/DashboardPage'
import { WorkspaceSettingsPage } from './pages/WorkspaceSettingsPage'
import { ContactsPage } from './pages/ContactsPage'
import { createRouter } from '@tanstack/react-router'

export interface ContactsSearch {
  cursor?: string
  email?: string
  externalId?: string
  firstName?: string
  lastName?: string
  phone?: string
  country?: string
  limit?: number
}

// Create the root route
const rootRoute = createRootRoute({
  component: RootLayout
})

// Create the index route
const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage
})

// Create the signin route
const signinRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/signin',
  component: SignInPage
})

// Create the logout route
const logoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/logout',
  component: LogoutPage
})

// Create the accept invitation route
const acceptInvitationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/accept-invitation',
  component: () => <div>Accept Invitation</div>
})

// Create the workspace create route
const workspaceCreateRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/workspace/create',
  component: CreateWorkspacePage
})

// Create the workspace route
const workspaceRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/workspace/$workspaceId',
  component: WorkspaceLayout
})

// Create workspace child routes
const workspaceCampaignsRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/campaigns',
  component: () => <div>Campaigns</div>
})

export const contactsRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/contacts',
  component: ContactsPage,
  validateSearch: (search: Record<string, unknown>): ContactsSearch => ({
    cursor: search.cursor as string | undefined,
    email: search.email as string | undefined,
    externalId: search.externalId as string | undefined,
    firstName: search.firstName as string | undefined,
    lastName: search.lastName as string | undefined,
    phone: search.phone as string | undefined,
    country: search.country as string | undefined,
    limit: search.limit ? Number(search.limit) : 20
  })
})

const workspaceSettingsRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/settings',
  component: WorkspaceSettingsPage
})

const workspaceTemplatesRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/templates',
  component: () => <div>Templates</div>
})

// Create the router
const routeTree = rootRoute.addChildren([
  indexRoute,
  signinRoute,
  logoutRoute,
  acceptInvitationRoute,
  workspaceCreateRoute,
  workspaceRoute.addChildren([
    workspaceCampaignsRoute,
    contactsRoute,
    workspaceSettingsRoute,
    workspaceTemplatesRoute
  ])
])

// Create and export the router with explicit type
export const router = createRouter({
  routeTree
})

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
