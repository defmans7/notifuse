import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/workspace/$workspaceId/campaigns')({
  component: () => <div>Campaigns page coming soon</div>
})
