import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/workspace/$workspaceId/contacts')({
  component: () => <div>Contacts page coming soon</div>
})
