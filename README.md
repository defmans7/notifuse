# Notifuse

[![Go Report Card](https://goreportcard.com/badge/github.com/Notifuse/notifuse)](https://goreportcard.com/report/github.com/Notifuse/notifuse)
[![Go](https://github.com/Notifuse/notifuse/actions/workflows/go.yml/badge.svg)](https://github.com/Notifuse/notifuse/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/Notifuse/notifuse/graph/badge.svg?token=VZ0HBEM9OZ)](https://codecov.io/gh/Notifuse/notifuse)

Notifuse is an email marketing platform that allows you to send emails to your contacts.

## Main Features

- Users can create and manage their own workspaces
- Send emails to your contacts
- Create and manage your contacts
- Create and manage subscription lists
- Create and manage your email templates with drag and drop editor
- Create and manage your email campaigns

## Project Structure

- `/server`: Backend Golang API
- `/console`: Frontend Vite + React + Ant Design + TypeScript source for the main application
- `/server/console`: Built frontend files served by the Golang API (generated during build process)

## Components

1. **API Server**: Golang based backend that handles core functionality and serves the console.
2. **Console**: Admin interface for managing workspaces, notifications, and settings.

## Technologies Used

- Backend: Golang
- Frontend (Console): Vite, React, Ant Design, TypeScript, Tanstack Router, Tailwind CSS, Tanstack Query
- Deployment: Docker
- Database: PostgreSQL

## Build and Deployment

Notifuse is designed for easy deployment:

1. During the build process, the frontend console is built and its files are copied to the `/server/console` folder.
2. The Golang server is configured to serve these static files alongside the API.
3. The entire application (both frontend and backend) is containerized into a single Docker image.

This approach allows for simple deployment and scaling of the entire Notifuse application as a single unit, while maintaining a separate, easily updatable marketing presence.

## License

Notifuse is released under the [Elastic License 2.0](LICENSE).

## Contact

For support or inquiries, please contact us at [hello@notifuse.com] or visit our [website](https://www.notifuse.com).
