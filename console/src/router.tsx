import { createRootRoute, createRoute } from '@tanstack/react-router'
import { RootLayout } from './layouts/RootLayout'
import { WorkspaceLayout } from './layouts/WorkspaceLayout'
import { SignInPage } from './pages/SignInPage'
import { LogoutPage } from './pages/LogoutPage'
import { CreateWorkspacePage } from './pages/CreateWorkspacePage'
import { DashboardPage } from './pages/DashboardPage'
import { createRouter } from '@tanstack/react-router'

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

const workspaceContactsRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/contacts',
  component: () => <div>Contacts</div>
})

const workspaceSettingsRoute = createRoute({
  getParentRoute: () => workspaceRoute,
  path: '/settings',
  component: () => <div>Settings</div>
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
    workspaceContactsRoute,
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
