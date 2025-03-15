# Workspace Context

## Properties

- `name`: Human-readable workspace name
- `id`: Short identifier compatible with PostgreSQL database naming
- `website_url`: URL to the workspace's website (optional)
- `logo_url`: URL to the workspace's logo (optional)
- `created_at`: Timestamp of workspace creation
- `updated_at`: Timestamp of last workspace update
- `timezone`: Workspace's timezone (e.g., "America/New_York")

## Multi-tenancy

- Each workspace has its own dedicated database for complete tenant isolation
- Database name is derived from the workspace's `id` and contain the prefixes `ntf_ws_`

## User Management

- A workspace can have multiple users
- Each user has one of two roles:
  - **Owner**: Full administrative control (only one per workspace)
  - **Member**: Standard workspace access

## Owner Capabilities

- Invite new members to the workspace
- Remove existing members from the workspace
- Transfer workspace ownership to another member
  - After transfer, original owner becomes a member
  - Only one owner exists at any time
