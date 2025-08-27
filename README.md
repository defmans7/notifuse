# Notifuse

[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/Notifuse/notifuse)
[![Go](https://github.com/Notifuse/notifuse/actions/workflows/go.yml/badge.svg)](https://github.com/Notifuse/notifuse/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/Notifuse/notifuse/graph/badge.svg?token=VZ0HBEM9OZ)](https://codecov.io/gh/Notifuse/notifuse)

**[🎯 Try the Live Demo](https://demo.notifuse.com/signin?email=demo@notifuse.com)**

**The open-source alternative to Mailchimp, Brevo, Mailjet, Listmonk, Mailerlite, and Klaviyo, Loop.so, etc.**

Notifuse is a modern, self-hosted emailing platform that allows you to send newsletters and transactional emails at a fraction of the cost. Built with Go and React, it provides enterprise-grade features with the flexibility of open-source software.

<img src="https://www.notifuse.com/_astro/email_editor.CGyLoCOD.png" alt="Email Editor">

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
- **Multi-Provider Support**: Connect with Amazon SES, SendGrid, Mailgun, Postmark, Mailjet, SparkPost, and SMTP

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
└── config/                # Configuration files
```

## 🚀 Getting Started

### Quick Start with Docker Compose

1. **Clone the repository**:

   ```bash
   git clone https://github.com/Notifuse/notifuse.git
   cd notifuse
   ```

2. **Configure environment variables**:

   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Generate PASETO keys**:
   Visit [paseto.notifuse.com](https://paseto.notifuse.com) to generate your PASETO keys, or use the built-in keygen command:

   ```bash
   # Generate keys manually
   make keygen

   # Or directly with Go
   go run cmd/keygen/main.go
   ```

4. **Start the services**:

   ```bash
   docker-compose up -d
   ```

5. **Access the application**:
   - Console: http://localhost:8080
   - API: http://localhost:8080/api

### Environment Configuration

**⚠️ Important**: The included `docker-compose.yml` is designed for **testing and development only**. For production deployments:

- **Use a separate PostgreSQL database** (managed service recommended)
- **Configure external storage** for file uploads
- **Set up proper SSL/TLS termination**
- **Use a reverse proxy** (nginx, Traefik, etc.)

#### Development Setup

The docker-compose includes a PostgreSQL container for quick testing. Simply run `docker-compose up -d` to get started.

#### Production Setup

Copy `env.example` to `.env` and configure:

- **Required**: `ROOT_EMAIL`, `API_ENDPOINT`, `PASETO_PRIVATE_KEY`, `PASETO_PUBLIC_KEY`
- **Database**: Configure `DB_HOST`, `DB_USER`, `DB_PASSWORD` for your external PostgreSQL
- **SMTP Settings**: Configure your email provider for system emails
- **SSL**: Set `DB_SSLMODE=require` for secure database connections

For detailed installation instructions, configuration options, and setup guides, visit **[docs.notifuse.com](https://docs.notifuse.com)**.

## 📚 Documentation

- **[Complete Documentation](https://docs.notifuse.com)** - Comprehensive guides and tutorials

## 🤝 Contributing

We welcome contributions!

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

Notifuse is released under the [GNU Affero General Public License v3.0](LICENSE).

## 🆘 Support

- **Documentation**: [docs.notifuse.com](https://docs.notifuse.com)
- **Email Support**: [hello@notifuse.com](mailto:hello@notifuse.com)
- **GitHub Issues**: [Report bugs or request features](https://github.com/Notifuse/notifuse/issues)

## 🌟 Why Choose Notifuse?

- **💰 Cost-Effective**: Self-hosted solution with no per-email pricing
- **🔒 Privacy-First**: Your data stays on your infrastructure
- **🛠️ Customizable**: Open-source with extensive customization options
- **📈 Scalable**: Built to handle millions of emails
- **🚀 Modern**: Built with modern technologies and best practices
- **🔧 Developer-Friendly**: Comprehensive API and webhook support

---

**Ready to get started?** [Try the live demo](https://demo.notifuse.com/signin?email=demo@notifuse.com) or [deploy your own instance](https://docs.notifuse.com) in minutes.
