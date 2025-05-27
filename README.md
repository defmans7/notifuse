# Notifuse

[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/Notifuse/notifuse)
[![Go](https://github.com/Notifuse/notifuse/actions/workflows/go.yml/badge.svg)](https://github.com/Notifuse/notifuse/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/Notifuse/notifuse/graph/badge.svg?token=VZ0HBEM9OZ)](https://codecov.io/gh/Notifuse/notifuse)

**The open-source alternative to Mailchimp, Brevo, Mailjet, Listmonk, Mailerlite, and Klaviyo.**

Notifuse is a modern, self-hosted email marketing platform that allows you to send newsletters and transactional emails at a fraction of the cost. Built with Go and React, it provides enterprise-grade features with the flexibility of open-source software.

## 🚀 Key Features

### 📧 Email Marketing

- **Visual Email Builder**: Drag-and-drop editor with MJML components and real-time preview
- **Campaign Management**: Create, schedule, and send targeted email campaigns
- **A/B Testing**: Optimize campaigns with built-in testing for subject lines, content, and send times
- **List Management**: Advanced subscriber segmentation and list organization
- **Contact Profiles**: Rich contact management with custom fields and detailed profiles

### 🔧 Developer-Friendly

- **Transactional API**: Powerful REST API for automated email delivery
- **Webhook Integration**: Real-time event notifications and integrations
- **Liquid Templating**: Dynamic content with variables like `{{ contact.first_name }}`
- **Multi-Provider Support**: Connect with SendGrid, Mailgun, SES, Postmark, Mailjet, SparkPost, and SMTP

### 📊 Analytics & Insights

- **Open & Click Tracking**: Detailed engagement metrics and campaign performance
- **Real-time Analytics**: Monitor delivery rates, opens, clicks, and conversions
- **Campaign Reports**: Comprehensive reporting and analytics dashboard

### 🎨 Advanced Features

- **S3 File Manager**: Integrated file management with CDN delivery
- **Notification Center**: Centralized notification system for your applications
- **Responsive Templates**: Mobile-optimized email templates
- **Custom Fields**: Flexible contact data management
- **Workspace Management**: Multi-tenant support for teams and agencies

## 🏗️ Architecture

Notifuse follows clean architecture principles with clear separation of concerns:

### Backend (Go)

- **Domain Layer**: Core business logic and entities (`internal/domain/`)
- **Service Layer**: Business logic implementation (`internal/service/`)
- **Repository Layer**: Data access and storage (`internal/repository/`)
- **HTTP Layer**: API handlers and middleware (`internal/http/`)

### Frontend (React)

- **Console**: Admin interface built with React, Ant Design, and TypeScript (`console/`)
- **Notification Center**: Embeddable widget for customer notifications (`notification_center/`)

### Database

- **PostgreSQL**: Primary data storage with Squirrel query builder
- **Migrations**: Database schema management and versioning

## 📁 Project Structure

```
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── domain/            # Business entities and logic
│   ├── service/           # Business logic implementation
│   ├── repository/        # Data access layer
│   ├── http/              # HTTP handlers and middleware
│   └── database/          # Database configuration
├── console/               # React-based admin interface
├── notification_center/   # Embeddable notification widget
├── pkg/                   # Public packages
├── docs/                  # Documentation (Mintlify)
├── homepage/              # Marketing website (Astro)
└── config/                # Configuration files
```

## 🚀 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/Notifuse/notifuse.git
cd notifuse

# Start with Docker Compose
docker-compose up -d

# Access the console
open http://localhost:8080
```

### Manual Installation

```bash
# Install dependencies
go mod download
cd console && npm install

# Build the application
make build

# Run the server
./bin/notifuse
```

## 🔧 Configuration

Notifuse can be configured through environment variables or configuration files:

```env
# Database
DATABASE_URL=postgres://user:password@localhost/notifuse

# Email Providers
SENDGRID_API_KEY=your_sendgrid_key
MAILGUN_API_KEY=your_mailgun_key
AWS_ACCESS_KEY_ID=your_aws_key

# File Storage
S3_BUCKET=your-bucket-name
S3_REGION=us-east-1

# Application
APP_PORT=8080
APP_ENV=production
```

## 📚 Documentation

- **[Complete Documentation](https://docs.notifuse.com)** - Comprehensive guides and tutorials
- **[API Reference](https://docs.notifuse.com/api-reference)** - REST API documentation
- **[Self-Hosting Guide](https://docs.notifuse.com/deployment)** - Deployment and configuration
- **[Developer Guide](https://docs.notifuse.com/development)** - Contributing and customization

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](https://docs.notifuse.com/development/contributing) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

Notifuse is released under the [Elastic License 2.0](LICENSE).

## 🆘 Support

- **Documentation**: [docs.notifuse.com](https://docs.notifuse.com)
- **Email Support**: [hello@notifuse.com](mailto:hello@notifuse.com)
- **GitHub Issues**: [Report bugs or request features](https://github.com/Notifuse/notifuse/issues)
- **Community**: [Join our Discord](https://discord.gg/notifuse)

## 🌟 Why Choose Notifuse?

- **💰 Cost-Effective**: Self-hosted solution with no per-email pricing
- **🔒 Privacy-First**: Your data stays on your infrastructure
- **🛠️ Customizable**: Open-source with extensive customization options
- **📈 Scalable**: Built to handle millions of emails
- **🚀 Modern**: Built with modern technologies and best practices
- **🔧 Developer-Friendly**: Comprehensive API and webhook support

---

**Ready to get started?** [Try the live demo](https://demo.notifuse.com) or [deploy your own instance](https://docs.notifuse.com/deployment/docker) in minutes.
