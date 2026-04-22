# Blinky

**A zero-configuration, self-managed headless CMS engine powered by Go & PostgreSQL.**

![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16.3-336791?style=for-the-badge&logo=postgresql)
![React](https://img.shields.io/badge/React_19-Vite-61DAFB?style=for-the-badge&logo=react)
![Alpha](https://img.shields.io/badge/Alpha-0.1.1-61DAFB?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)
![Cross-Platform](https://img.shields.io/badge/Platform-Win%20%7C%20Mac%20%7C%20Linux-lightgrey?style=for-the-badge)

> **Experimental Project**
> Blinky is a hobby project built entirely through vibe coding with AI assistance. It is **not intended for production use** in professional or enterprise environments. I may suspend or discontinue this project at any time — I cannot be held responsible for any issues arising from serious use.

---

## What is Blinky?

Blinky puts PostgreSQL at its core and makes **dynamic schema management** as simple and fast as possible. Instead of wrestling with complex table structures and API boilerplate, you define your data, and Blinky instantly gives you a polished admin dashboard and ready-to-use REST API endpoints.

Blinky **auto-manages its own PostgreSQL instance** — it downloads, initializes, and runs a dedicated PostgreSQL server with zero manual configuration. Just run the binary and start building.

---

## Features

### Core Engine
- **Zero-Configuration PostgreSQL:** Blinky automatically downloads, initializes, and manages its own PostgreSQL 16.3 instance. No manual database setup required.
- **Fully Modular SQL Architecture:** Zero hard-coded SQL strings anywhere. DDL via the Statement Builder, CRUD via the Squirrel Query Builder.
- **Dynamic Collection Management:** Define tables from the admin panel. Blinky updates the database schema and exposes API routes instantly.
- **Sequential Database Migrations:** Version-controlled, transactional schema migrations with automatic pre-migration backups.

### Security
- **Isolated Services:** The Admin Panel and Public API run on independent hosts/ports — admin stays on `localhost`, public API is exposed externally.
- **Argon2id Password Hashing:** Modern, memory-hard password hashing with automatic legacy bcrypt migration support.
- **Hardened Sessions:** `httpOnly`, `Secure`, `SameSite` cookie flags with CSRF protection on all mutating endpoints.
- **SSH Tunnel Gateway:** Secure remote access to admin and public API services via encrypted SSH port forwarding.
- **Smart Setup Flow:** The system remains locked until the first administrator account is created.

### Admin Dashboard
- **Premium UI:** Built with React 19, Vite, and Shadcn UI — fast, responsive, and premium-looking.
- **Responsive DataTables:** Intelligent "Vertical View" card layout for all tables on mobile.
- **SQL Query Editor:** Integrated web-based SQL editor with PrismJS syntax highlighting, transaction support, and real-time result rendering.
- **Backup & Restore:** One-click database backups with `pg_dump` compression and download/delete management.
- **PostgreSQL Config Editor:** Edit `postgresql.conf` parameters directly through the admin panel with live toggling support.
- **Environment Manager:** View, edit, and delete `.env` variables from the dashboard.
- **Server Settings:** Configure API hosts, ports, and SSH tunnel settings from the UI.

### Cross-Platform
- Full support for **Linux**, **macOS**, and **Windows**.
- Path handling via `pathutil` abstraction — no hardcoded separators.
- OS-agnostic system commands throughout the codebase.

---

## Supported Field Types

Blinky collections (tables) support a rich set of field types:

| Type | Description |
|---|---|
| **ID** | Auto-generated unique identifiers with configurable length |
| **Text** | String content with optional min/max length constraints |
| **Number** | Integer or decimal values with min/max/no-zero/no-decimals options |
| **Boolean** | `true/false` toggles with configurable defaults |
| **Date** | Timestamps with auto-managed `created_at` / `updated_at` support |
| **JSON** | Flexible JSONB structures without a fixed schema |
| **Relation** | Cross-collection references (single or multiple mode) |

---

## Tech Stack

### Backend

| Technology | Role |
|---|---|
| [Go 1.26](https://go.dev/) | Core language |
| [PostgreSQL 16.3](https://www.postgresql.org/) | Auto-managed database |
| [Fiber v2](https://gofiber.io/) | Web framework |
| [PGX v5](https://github.com/jackc/pgx) | PostgreSQL driver |
| [Squirrel](https://github.com/Masterminds/squirrel) | SQL query builder |
| [Argon2id](https://pkg.go.dev/golang.org/x/crypto/argon2) | Password hashing |
| [x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh) | SSH tunnel gateway |
| [google/uuid](https://github.com/google/uuid) | Deterministic field IDs |

### Frontend

| Technology | Role |
|---|---|
| [React 19](https://react.dev/) & [Vite](https://vitejs.dev/) | UI framework & build tool |
| [TailwindCSS](https://tailwindcss.com/) | Styling |
| [Shadcn UI](https://ui.shadcn.com/) | Component library |
| [Lucide React](https://lucide.dev/) | Icons |
| [PrismJS](https://prismjs.com/) | SQL syntax highlighting |

---

## Project Structure

```
blinky/
├── internal/
│   ├── api/
│   │   ├── admin/
│   │   │   ├── auth/          # Authentication & session management
│   │   │   ├── collections/   # Collection CRUD & schema management
│   │   │   │   └── tables/    # System table schema definitions
│   │   │   ├── settings/      # Backup, env, postgres, server config
│   │   │   └── system/        # Engine control, file browser, SQL query
│   │   ├── public/            # Public REST API endpoints
│   │   ├── errors.go          # Centralized error constants
│   │   ├── success.go         # Centralized success messages
│   │   ├── response.go        # Unified response envelope
│   │   ├── utils.go           # Shared helpers (ID gen, DB error handling)
│   │   └── validator.go       # Custom input validation
│   ├── config/                # Environment & configuration management
│   ├── database/              # SQL builders, migrations, keywords
│   ├── panel/                 # React/Vite admin dashboard (embedded)
│   ├── pkg/
│   │   ├── crypto/            # Argon2id password hashing
│   │   ├── logger/            # Centralized color-coded logger
│   │   ├── pathutil/          # Cross-platform path utilities
│   │   ├── postgresql/        # Auto-managed PG lifecycle & conf parser
│   │   ├── ssh/               # SSH tunnel gateway server
│   │   └── worker/            # Background task manager
│   └── types/                 # Core data structures (CollectionSchema, etc.)
├── scripts/                   # Development utilities
├── main.go                    # Application entry point
├── go.mod
└── .env                       # Runtime configuration
```

---

## Getting Started

### Prerequisites

- **Go** v1.26+
- **Node.js** v24+ LTS *(only required to build the admin panel)*

> **Note:** PostgreSQL is **not** required as a prerequisite — Blinky downloads and manages its own instance automatically.

---

### Quick Start

**1. Download Blinky**

Grab a pre-built binary from the [Releases](https://github.com/itsmelissadev/blinky/releases) page for your OS.

Or clone and build manually:

```bash
git clone https://github.com/itsmelissadev/blinky.git
cd blinky
```

**2. Run the Project**

```bash
go mod tidy
go run main.go
```

On first run, Blinky will:
1. Download and initialize a managed PostgreSQL 16.3 instance
2. Prompt you for secure database credentials
3. Run schema migrations automatically
4. Start both the Admin Panel (`:8080`) and Public API (`:8090`)

> Navigate to `http://localhost:8080` to access the **Initial Setup Wizard** and create your first admin account.

---

### Development Setup

For live development with hot reload:

```bash
# Terminal 1: Start the Go backend with Air
air -c .air.toml

# Terminal 2: Start the Vite dev server for the admin panel
cd internal/panel
npm install
npm run dev
```

**Development Rules:**
- All SQL queries must use the `Squirrel` builder — raw SQL strings are prohibited.
- All schema operations must use the `Statement` builder from `internal/database`.
- API errors go in `errors.go`, success messages in `success.go` — no hardcoded strings.
- Reusable functions must be centralized in `pkg/` or `utils.go`.

---

## API Overview

### Admin API (`/_api/`)

| Method | Endpoint | Description |
|---|---|---|
| GET | `/_api/admins/initialized` | Check if setup is complete |
| POST | `/_api/admins/login` | Admin login |
| POST | `/_api/admins/user` | Create admin account |
| GET | `/_api/admins/me` | Get current admin |
| GET | `/_api/collections/` | List all collections |
| POST | `/_api/collection/` | Create a collection |
| PATCH | `/_api/collection/:name` | Update a collection |
| DELETE | `/_api/collection/:name` | Delete a collection |
| GET | `/_api/collection/:name/records` | List records |
| POST | `/_api/collection/:name/records` | Create a record |
| PATCH | `/_api/collection/:name/records/:id` | Update a record |
| POST | `/_api/system/sql` | Execute raw SQL query |
| GET | `/_api/settings/backup/` | List backups |
| POST | `/_api/settings/backup/` | Create backup |

### Public API (`:8090`)

| Method | Endpoint | Description |
|---|---|---|
| GET | `/collections/:name` | List records with filtering & sorting |
| POST | `/collections/:name` | Create a record |
| GET | `/collections/:name/:id` | Get a single record |
| PATCH | `/collections/:name/:id` | Update a record |
| DELETE | `/collections/:name/:id` | Delete a record |

**Query Operators:** `eq`, `neq`, `gt`, `lt`, `gte`, `lte`, `like`, `in`
**Sorting:** `?sort=field` (ascending) or `?sort=-field` (descending)
**Pagination:** `?limit=100&offset=0`

---

## Contributing

Blinky is an open-source project and welcomes all feedback. If you encounter a bug or have a suggestion for a new feature, feel free to open an **Issue**.

> **Note:** Due to our current workload, we are temporarily unable to accept pull requests; however, your suggestions and ideas are always welcome and will be carefully reviewed!

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
