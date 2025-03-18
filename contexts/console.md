The console is a frontend built with React, Ant Design, Tanstack Routerand Tanstack Query.

- The root route is `/` and is the main entry point for the console.
- The console is protected and requires a user to be authenticated to access any route, except for public routes `/signin` and `/accept-invitation` and `/logout`.
- If the user is not authenticated, they are redirected to the `/signin` route.
- If the user is authenticated, they are redirected to the `/` route.
- The `/workspace/create` route is used to create a new workspace.
- If the user has no workspaces, they are redirected to the `/workspace/create` route.
- The `/logout` route is used to logout the user and redirect to the `/signin` route.
- The `/accept-invitation` route is used to accept an invitation to a workspace and redirect to the `/` route.
- The `/workspace/:workspaceId` route is used to access a specific workspace.
- The `/workspace/:workspaceId/campaigns` route is used to access the campaigns of a specific workspace.
- On signin, the workspaces are fetched from the backend and stored in the `AuthContext`.
- Inside a workspace, a layout is used with a sidebar and a content area.
- The sidebar contains the navigation menu with the following items:
  - `Campaigns`
  - `Contacts`
  - `Settings`
  - `Templates`
  - `Logout`
