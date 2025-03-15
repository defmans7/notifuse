## Database Structure

### System Database

The system database is named `ntf_system` and contains:

- `users`: User accounts and authentication
- `workspaces`: Workspace configuration and metadata
- `user_workspaces`: Many-to-many relationship between users and workspaces
- `settings`: System-wide configuration settings
- `user_sessions`: Active user session data

### Workspace-specific Database

Each workspace maintains its own separate database containing:

- Database name is derived from the workspace's `id` and contain the prefixes `ntf_ws_`
- `contacts`: Contact information and metadata
- `subscription_lists`: Mailing lists and subscriber groups
- `templates`: Email and message templates
- `campaigns`: Marketing campaign data and configuration
- etc...
