# Delivery & Household Services Backend API

Production-ready, high-performance RESTful backend API built with **Go (Golang)**, **PostgreSQL 16**, and **Docker Compose**, following industry-standard **Clean Architecture** patterns.

---

## 🏗 Architecture & Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Application entrypoint & server lifecycle
├── internal/
│   ├── config/                     # Environment configuration loader
│   ├── database/                   # PostgreSQL connection & auto-seeders
│   ├── dto/                        # Data Transfer Objects & validation
│   ├── handler/                    # HTTP controllers (REST endpoints)
│   ├── middleware/                 # JWT Auth, CORS, Logger, Recovery
│   ├── models/                     # GORM domain models & database schemas
│   ├── repository/                 # Data access layer (CRUD & queries)
│   ├── service/                    # Business logic implementation
│   └── utils/                      # Password hashing, JWT, Standard JSON responses
├── .dockerignore
├── .env.example
├── .env
├── Dockerfile                      # Multi-stage optimized Docker build
├── docker-compose.yml              # PostgreSQL + Go API orchestrated in Docker
├── Makefile                        # Handy shortcut commands
└── README.md
```

---

## 🚀 Quick Start with Docker

### 1. Prerequisites
- Docker (version 20.10+)
- Docker Compose (v2+)

### 2. Start Services
From the `backend` directory, run:

```bash
docker compose up -d --build
```

Or using `make`:
```bash
make build
```

This starts:
- **PostgreSQL 16**: Port `5432` with automatic health checks & persistent data volume
- **Go API**: Port `8080` with auto-migration, seeding, and health checks

### 3. Check Status & Logs
```bash
docker compose ps
docker compose logs -f api
```

### 4. Stop Services
```bash
docker compose down
```

---

## 🔑 Pre-seeded Test Accounts & Data

- **Customer Demo Account:**
  - **Email:** `user@example.com`
  - **Password:** `Password123!`
  - **Phone:** `+94 77 123 4567`
- **Pre-seeded Services:** Plumbing, Electricity, Building Work, Cleaning, Care Taking, Carpenter, Cook & Chef, Other Services
- **Pre-seeded Workers:** 7 verified professional profiles with ratings, reviews, and badges in Sri Lanka
- **Pre-seeded Promo Codes:** `WELCOME20` (20% off up to Rs. 1,500), `SAVE500` (Flat Rs. 500 off)

---

## 📚 API Endpoints Reference

Base URL: `http://localhost:8080/api/v1`

### 1. System & Health
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/` | API Information | No |
| `GET` | `/health` | Healthcheck & Database status | No |

### 2. Authentication
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | Register new customer account | No |
| `POST` | `/api/v1/auth/login` | Login with email/phone & password | No |

#### Example Login Request:
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone_or_email": "user@example.com",
    "password": "Password123!"
  }'
```

### 3. User Profile & Addresses
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/user/profile` | Get current user profile | **Yes (Bearer)** |
| `PUT` | `/api/v1/user/profile` | Update profile information | **Yes (Bearer)** |
| `POST` | `/api/v1/user/change-password` | Change user password | **Yes (Bearer)** |
| `GET` | `/api/v1/user/locations` | List saved addresses | **Yes (Bearer)** |
| `POST` | `/api/v1/user/locations` | Save a new address | **Yes (Bearer)** |
| `PUT` | `/api/v1/user/locations/:id/default` | Set default address | **Yes (Bearer)** |
| `DELETE` | `/api/v1/user/locations/:id` | Delete saved address | **Yes (Bearer)** |

### 4. Service Catalog
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/services` | List all available services & sub-options | No |
| `GET` | `/api/v1/services/:id` | Get service detail by UUID or code | No |

### 5. Workers / Service Specialists
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/workers` | List workers (optional `?service_type=cleaning`) | No |
| `GET` | `/api/v1/workers/:id` | Get worker profile & reviews | No |
| `POST` | `/api/v1/workers/:id/reviews` | Submit rating & review for worker | **Yes (Bearer)** |

### 6. Vouchers & Discounts
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `GET` | `/api/v1/vouchers` | List active promotional vouchers | No |
| `POST` | `/api/v1/vouchers/validate` | Validate code & calculate discount | No |

### 7. Bookings & Orders
| Method | Endpoint | Description | Auth Required |
|---|---|---|---|
| `POST` | `/api/v1/bookings` | Create a new service booking | **Yes (Bearer)** |
| `GET` | `/api/v1/bookings` | List authenticated user bookings | **Yes (Bearer)** |
| `GET` | `/api/v1/bookings/:id` | Get booking detail | **Yes (Bearer)** |
| `POST` | `/api/v1/bookings/:id/cancel` | Cancel booking | **Yes (Bearer)** |
| `PATCH` | `/api/v1/bookings/:id/status` | Update booking & payment status | **Yes (Bearer)** |

#### Example Create Booking Request:
```bash
curl -X POST http://localhost:8080/api/v1/bookings \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "<SERVICE_UUID>",
    "worker_id": "<WORKER_UUID>",
    "location_title": "The Base Sukhumvit 77",
    "location_subtitle": "The Base Sukhumvit 77, On Nut Rd, Bangkok",
    "total_cost": 2687.00,
    "voucher_code": "WELCOME20",
    "payment_method": "promptpay",
    "notes": "Deep clean 2 wall AC units and condenser wash",
    "selected_items": {
      "Wall-Mounted AC": 2,
      "Condenser External Wash": 1
    }
  }'
```

---

## 🔒 Security Best Practices Implemented

1. **Password Hashing:** Passwords encrypted using **bcrypt** with secure cost factor.
2. **JWT Authentication:** Stateless, signed JWT Bearer tokens with configurable expiration.
3. **Multi-stage Docker Build:** Minimal Alpine Linux image runner with **non-root user** (`appuser:appgroup`).
4. **Resilient Database Lifecycles:** Automatic retry loop for DB connectivity during container orchestration.
5. **SQL Injection Prevention:** GORM parameterized ORM queries across all data layers.
6. **Graceful Shutdown:** Handles OS termination signals (`SIGINT`, `SIGTERM`) without dropping active connections.
