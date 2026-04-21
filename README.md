# Blinky

**A dynamic database management system built with Go, PostgreSQL, and React (Vite).**

![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18-336791?style=for-the-badge&logo=postgresql)
![React](https://img.shields.io/badge/React-Vite-61DAFB?style=for-the-badge&logo=react)
![Alpha](https://img.shields.io/badge/Alpha-0.1.0-61DAFB?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)

> **Experimental Project**
> Blinky is a hobby project built entirely through vibe coding with AI assistance. It is **not intended for production use** in professional or enterprise environments. I may suspend or discontinue this project at any time - I cannot be held responsible for any issues arising from serious use.

---

## What is Blinky?

Blinky puts PostgreSQL at its core and makes **dynamic schema management** as simple and fast as possible. Instead of wrestling with complex table structures and API boilerplate, you define your data, and Blinky instantly gives you a polished admin dashboard and ready-to-use REST API endpoints. All in one place.

---

## Features

- **Fully Modular SQL Architecture:** There are no hard-coded SQL strings anywhere in the codebase. Database schemas are created using a dedicated Statement Builder, while all CRUD operations are executed securely and flexibly via the *Squirrel Query Builder*.
- **High-Performance:** It is built on low-latency, scalable RESTful endpoints.
- **Security:**
  - The `Admin Panel` and `Public API` run independently on different hosts and ports, so while the Admin Panel is accessible only via `localhost`, the Public API remains accessible from anywhere.
  - Smart setup mechanism: The system remains closed off from the outside world until the first administrator account is created.
- **Dynamic Collection Management:** Define new tables directly from the admin panel. Blinky updates the database schema and makes the relevant API routes available immediately.
- **Modern Admin Dashboard:** It was developed using React, Vite, and Shadcn UI; it is fast, responsive, and intuitive.
- **Backup and Restore:** Quickly backup and restore your entire database via the admin panel.
- **PostgreSQL Config Editor:** Edit the `postgresql.conf` file in PostgreSQL directly through the admin panel.

---

## Supported Field Types

Blinky collections (tables) support a rich set of field types:

| Type | Description |
|---|---|
| **Text** | Short or long string content |
| **Number** | Integer or decimal numeric values |
| **Boolean** | `true/false`, active/inactive toggles |
| **Date** | Timestamps and custom date fields |
| **JSON** | Complex, dynamic tree structures without a fixed schema |
| **Relation** | Cross-collection references and data associations |

---

## Tech Stack

### Backend

| Technology | Role |
|---|---|
| [Go 1.26](https://go.dev/) | Core language |
| [PostgreSQL 18](https://www.postgresql.org/) | Database |
| [Fiber v2](https://gofiber.io/) | Web framework |
| [PGX v5](https://github.com/jackc/pgx) | PostgreSQL driver |
| [Squirrel](https://github.com/Masterminds/squirrel) | SQL query builder |
| `google/uuid`, `x/crypto` | Utilities |

### Frontend

| Technology | Role |
|---|---|
| [React 16+](https://react.dev/) & [Vite](https://vitejs.dev/) | UI framework & build tool |
| [TailwindCSS](https://tailwindcss.com/) | Styling |
| [Shadcn UI](https://ui.shadcn.com/) | Component library |
| [Lucide React](https://lucide.dev/) | Icons |

---

## Getting Started

### Prerequisites

- **Go** v1.26+
- **Node.js** v24.15.0 LTS *(required to build the admin panel)*
- **PostgreSQL** v18

---

### Quick Start (For Users)

**1. Download Blinky**

The easiest way is to grab a pre-built binary from the [Releases](https://github.com/itsmelissadev/blinky/releases) page for your OS.

Alternatively, clone the repository to build manually:

```bash
git clone https://github.com/itsmelissadev/blinky.git
cd blinky
```

**2. Set Up PostgreSQL**

Make sure PostgreSQL 18 is installed and running, then create an empty database for Blinky.

**If you have other important databases on PostgreSQL, I recommend testing in a separate PostgreSQL environment. I don’t want your other databases to be affected in case any errors occur, since the project isn’t yet in a usable state.**

**3. Run the Project**

```bash
# Install Go dependencies and start the backend
go mod tidy
go run main.go

# In a separate terminal, start the admin panel
cd internal/panel
npm install
npm run dev
```

> Once the admin panel has loaded in your browser, you will be automatically redirected to the **Initial Setup Wizard**.

---

### Developer Setup

If you want to contribute, fork, or customize Blinky:

1. **Fork the repo:** Work from your own fork on GitHub.
2. **Live Reload with Air:** To reload Go changes instantly during development, use the `air` command:

```bash
   air -c .air.toml
```

3. **Apply the Modular SQL Policy:** All new queries must use the `Squirrel` builder; all table/schema operations must use the `Statement` builder. Raw SQL strings must not be used.
4. **Vite HMR:** To get full Hot Module Replacement support during frontend development, work in the `internal/panel` directory.

---

## Contributing

Blinky is an open-source project and welcomes all feedback. If you encounter a bug or have a suggestion for a new feature, feel free to open an **Issue**.

> **Note:** Due to our current workload, we are temporarily unable to accept pull requests; however, your suggestions and ideas are always welcome and will be carefully reviewed!

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
